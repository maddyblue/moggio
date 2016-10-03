package webm

import (
	"log"
	"time"

	"github.com/ebml-go/webm"
	"github.com/xlab/opus-go/opus"
	"github.com/xlab/vorbis-go/decoder"
	"github.com/xlab/vorbis-go/vorbis"
)

type Samples struct {
	Timecode        time.Duration
	Data            [][]float32
	DataInterleaved []float32
}

type ADecoder struct {
	closeC chan struct{}
	doneC  chan struct{}

	enabled    bool
	codec      ACodec
	channels   int
	sampleRate int

	src       <-chan webm.Packet
	voDSP     vorbis.DspState
	voBlock   vorbis.Block
	voPCM     [][][]float32
	opDecoder *opus.Decoder
	opPCM     []float32
}

type ACodec string

const (
	CodecVorbis ACodec = "A_VORBIS"
	CodecOpus   ACodec = "A_OPUS"
)

func NewADecoder(codec ACodec, codecPrivate []byte,
	channels, sampleRate int, src <-chan webm.Packet) *ADecoder {

	d := &ADecoder{
		closeC: make(chan struct{}),
		doneC:  make(chan struct{}),

		channels:   channels,
		sampleRate: sampleRate,
		codec:      codec,
		src:        src,
	}
	switch codec {
	case CodecVorbis:
		var info vorbis.Info
		vorbis.InfoInit(&info)
		var comment vorbis.Comment
		vorbis.CommentInit(&comment)
		err := decoder.ReadHeaders(codecPrivate, &info, &comment)
		if err != nil {
			log.Println("[WARN]", err)
			return nil
		}
		info.Deref()
		comment.Deref()
		if comment.Comments > 0 {
			comment.UserComments = make([][]byte, comment.Comments)
			comment.Deref()
			streamInfo := decoder.ReadInfo(&info, &comment)
			log.Println("vorbis:", streamInfo.Comments)
		}
		if int(info.Channels) != channels {
			d.channels = int(channels)
			log.Printf("[WARN] vorbis: channel count mismatch %d != %d", info.Channels, channels)
		}
		if int(info.Rate) != sampleRate {
			d.sampleRate = int(info.Rate)
			log.Printf("[WARN] vorbis: sample rate mismatch %d != %d", info.Rate, sampleRate)
		}
		ret := vorbis.SynthesisInit(&d.voDSP, &info)
		if ret != 0 {
			log.Println("[WARN] vorbis init:", ret)
		} else {
			d.voPCM = [][][]float32{
				make([][]float32, channels),
			}
			vorbis.BlockInit(&d.voDSP, &d.voBlock)
			d.enabled = true
		}
		return d
	case CodecOpus:
		var err int32
		d.opDecoder = opus.DecoderCreate(int32(sampleRate), int32(channels), &err)
		if err != opus.Ok {
			log.Println("[WARN] opus init:", err)
		} else {
			d.opPCM = make([]float32, samplesPerBuffer*channels)
			d.enabled = true
		}
		return d
	default:
		log.Println("[WARN] unsupported audio codec:", codec)
		return d
	}
}

func (a *ADecoder) Process(out chan<- Samples) {
	defer close(out)

	var firstTimecode time.Duration
	voFrame := make([][]float32, 0, samplesPerBuffer)
	opFrame := make([]float32, 0, samplesPerBuffer*a.channels)
	sendVoFrame := func() {
		out <- Samples{
			Data:     voFrame,
			Timecode: firstTimecode,
		}
		voFrame = make([][]float32, 0, samplesPerBuffer)
		firstTimecode = 0
	}
	sendOpFrame := func() {
		out <- Samples{
			DataInterleaved: opFrame,
			Timecode:        firstTimecode,
		}
		opFrame = make([]float32, 0, samplesPerBuffer*a.channels)
		firstTimecode = 0
	}
	defer func() {
		if len(voFrame) > 0 {
			sendVoFrame()
		}
		if len(opFrame) > 0 {
			sendOpFrame()
		}
	}()
	decodePkt := func(pkt *webm.Packet) {
		switch a.codec {
		case CodecVorbis:
			if len(pkt.Data) == 0 {
				return
			}
			packet := &vorbis.OggPacket{
				Packet: pkt.Data,
				Bytes:  len(pkt.Data),
			}
			ret := vorbis.Synthesis(&a.voBlock, packet)
			if ret == 0 {
				vorbis.SynthesisBlockin(&a.voDSP, &a.voBlock)

				samples := vorbis.SynthesisPcmout(&a.voDSP, a.voPCM)
				if samples == 0 {
					vorbis.SynthesisRead(&a.voDSP, samples)
					return
				}
				for ; samples > 0; samples = vorbis.SynthesisPcmout(&a.voDSP, a.voPCM) {
					space := int32(samplesPerBuffer - len(voFrame))
					if samples > space {
						samples = space
					}
					for i := 0; i < int(samples); i++ {
						sample := make([]float32, a.channels)
						for j := 0; j < a.channels; j++ {
							sample[j] = a.voPCM[0][j][:samples][i]
						}
						if firstTimecode == 0 {
							firstTimecode = pkt.Timecode
						}
						voFrame = append(voFrame, sample)
					}
					if len(voFrame) == samplesPerBuffer {
						sendVoFrame()
					}
					vorbis.SynthesisRead(&a.voDSP, samples)
				}
			} else {
				log.Println("[WARN] vorbis synthesis:", ret)
				return
			}
		case CodecOpus:
			sampleCount := opus.DecodeFloat(a.opDecoder, data(pkt.Data), int32(len(pkt.Data)),
				a.opPCM, samplesPerBuffer, 0)
			if sampleCount <= 0 {
				return
			}
			maxSpace := samplesPerBuffer * a.channels
			samples := a.opPCM[:int(sampleCount)*a.channels]
			for len(samples) > 0 {
				space := maxSpace - len(opFrame)
				if space > len(samples) {
					space = len(samples)
				}
				if firstTimecode == 0 {
					firstTimecode = pkt.Timecode
				}
				opFrame = append(opFrame, samples[:space]...)
				samples = samples[space:]
				if len(opFrame) == maxSpace {
					sendOpFrame()
				}
			}
		}
	}
	for {
		select {
		case <-a.closeC:
			close(a.doneC)
			return
		case pkt, ok := <-a.src:
			if !ok {
				close(a.doneC)
				return
			}
			if !a.enabled {
				continue
			}
			decodePkt(&pkt)
		}
	}
}

func (a *ADecoder) ResetPCM() {
	switch a.codec {
	case CodecVorbis:
		a.voPCM = [][][]float32{
			make([][]float32, a.channels),
		}
	case CodecOpus:
		a.opPCM = make([]float32, samplesPerBuffer*a.channels)
	}
}

func (a *ADecoder) Wait() {
	<-a.doneC
}

func (a *ADecoder) Close() {
	close(a.closeC)
}

func (a *ADecoder) Channels() int {
	return a.channels
}

func (a *ADecoder) SampleRate() int {
	return a.sampleRate
}
