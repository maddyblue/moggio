package dsound

type WaveFormatEx struct {
	FormatTag      WaveFormatTag
	Channels       uint16
	SamplesPerSec  uint32
	AvgBytesPerSec uint32
	BlockAlign     uint16
	BitsPerSample  uint16
	ExtSize        uint16
}

type WaveFormatExtensible struct {
	Format      WaveFormatEx
	Samples     uint16 // union { ValidBitsPerSample or SamplesPerBlock or Reserved }
	ChannelMask WFESpeaker
	SubFormat   GUID
}

type WaveFormatTag uint16

const (
	WAVE_FORMAT_UNKNOWN                 = WaveFormatTag(0x0000) // Microsoft Corporation
	WAVE_FORMAT_PCM                     = WaveFormatTag(0x0001) // Microsoft PCM format
	WAVE_FORMAT_MS_ADPCM                = WaveFormatTag(0x0002) // Microsoft ADPCM
	WAVE_FORMAT_IEEE_FLOAT              = WaveFormatTag(0x0003) // Micrososft 32 bit float format
	WAVE_FORMAT_VSELP                   = WaveFormatTag(0x0004) // Compaq Computer Corporation
	WAVE_FORMAT_IBM_CVSD                = WaveFormatTag(0x0005) // IBM Corporation
	WAVE_FORMAT_ALAW                    = WaveFormatTag(0x0006) // Microsoft Corporation
	WAVE_FORMAT_MULAW                   = WaveFormatTag(0x0007) // Microsoft Corporation
	WAVE_FORMAT_OKI_ADPCM               = WaveFormatTag(0x0010) // OKI
	WAVE_FORMAT_IMA_ADPCM               = WaveFormatTag(0x0011) // Intel Corporation
	WAVE_FORMAT_MEDIASPACE_ADPCM        = WaveFormatTag(0x0012) // Videologic
	WAVE_FORMAT_SIERRA_ADPCM            = WaveFormatTag(0x0013) // Sierra Semiconductor Corp
	WAVE_FORMAT_G723_ADPCM              = WaveFormatTag(0x0014) // Antex Electronics Corporation
	WAVE_FORMAT_DIGISTD                 = WaveFormatTag(0x0015) // DSP Solutions, Inc.
	WAVE_FORMAT_DIGIFIX                 = WaveFormatTag(0x0016) // DSP Solutions, Inc.
	WAVE_FORMAT_DIALOGIC_OKI_ADPCM      = WaveFormatTag(0x0017) // Dialogic Corporation
	WAVE_FORMAT_MEDIAVISION_ADPCM       = WaveFormatTag(0x0018) // Media Vision, Inc.
	WAVE_FORMAT_CU_CODEC                = WaveFormatTag(0x0019) // Hewlett-Packard Company
	WAVE_FORMAT_YAMAHA_ADPCM            = WaveFormatTag(0x0020) // Yamaha Corporation of America
	WAVE_FORMAT_SONARC                  = WaveFormatTag(0x0021) // Speech Compression
	WAVE_FORMAT_DSPGROUP_TRUESPEECH     = WaveFormatTag(0x0022) // DSP Group, Inc
	WAVE_FORMAT_ECHOSC1                 = WaveFormatTag(0x0023) // Echo Speech Corporation
	WAVE_FORMAT_AUDIOFILE_AF36          = WaveFormatTag(0x0024) // Audiofile, Inc.
	WAVE_FORMAT_APTX                    = WaveFormatTag(0x0025) // Audio Processing Technology
	WAVE_FORMAT_AUDIOFILE_AF10          = WaveFormatTag(0x0026) // Audiofile, Inc.
	WAVE_FORMAT_PROSODY_1612            = WaveFormatTag(0x0027) // Aculab plc
	WAVE_FORMAT_LRC                     = WaveFormatTag(0x0028) // Merging Technologies S.A.
	WAVE_FORMAT_DOLBY_AC2               = WaveFormatTag(0x0030) // Dolby Laboratories
	WAVE_FORMAT_GSM610                  = WaveFormatTag(0x0031) // Microsoft Corporation
	WAVE_FORMAT_MSNAUDIO                = WaveFormatTag(0x0032) // Microsoft Corporation
	WAVE_FORMAT_ANTEX_ADPCME            = WaveFormatTag(0x0033) // Antex Electronics Corporation
	WAVE_FORMAT_CONTROL_RES_VQLPC       = WaveFormatTag(0x0034) // Control Resources Limited
	WAVE_FORMAT_DIGIREAL                = WaveFormatTag(0x0035) // DSP Solutions, Inc.
	WAVE_FORMAT_DIGIADPCM               = WaveFormatTag(0x0036) // DSP Solutions, Inc.
	WAVE_FORMAT_CONTROL_RES_CR10        = WaveFormatTag(0x0037) // Control Resources Limited
	WAVE_FORMAT_NMS_VBXADPCM            = WaveFormatTag(0x0038) // Natural MicroSystems
	WAVE_FORMAT_ROLAND_RDAC             = WaveFormatTag(0x0039) // Roland
	WAVE_FORMAT_ECHOSC3                 = WaveFormatTag(0x003A) // Echo Speech Corporation
	WAVE_FORMAT_ROCKWELL_ADPCM          = WaveFormatTag(0x003B) // Rockwell International
	WAVE_FORMAT_ROCKWELL_DIGITALK       = WaveFormatTag(0x003C) // Rockwell International
	WAVE_FORMAT_XEBEC                   = WaveFormatTag(0x003D) // Xebec Multimedia Solutions Limited
	WAVE_FORMAT_G721_ADPCM              = WaveFormatTag(0x0040) // Antex Electronics Corporation
	WAVE_FORMAT_G728_CELP               = WaveFormatTag(0x0041) // Antex Electronics Corporation
	WAVE_FORMAT_MSG723                  = WaveFormatTag(0x0042) // Microsoft Corporation
	WAVE_FORMAT_MPEG                    = WaveFormatTag(0x0050) // Microsoft Corporation
	WAVE_FORMAT_RT24                    = WaveFormatTag(0x0052) // InSoft Inc.
	WAVE_FORMAT_PAC                     = WaveFormatTag(0x0053) // InSoft Inc.
	WAVE_FORMAT_MPEGLAYER3              = WaveFormatTag(0x0055) // MPEG 3 Layer 1
	WAVE_FORMAT_LUCENT_G723             = WaveFormatTag(0x0059) // Lucent Technologies
	WAVE_FORMAT_CIRRUS                  = WaveFormatTag(0x0060) // Cirrus Logic
	WAVE_FORMAT_ESPCM                   = WaveFormatTag(0x0061) // ESS Technology
	WAVE_FORMAT_VOXWARE                 = WaveFormatTag(0x0062) // Voxware Inc
	WAVE_FORMAT_CANOPUS_ATRAC           = WaveFormatTag(0x0063) // Canopus, Co., Ltd.
	WAVE_FORMAT_G726_ADPCM              = WaveFormatTag(0x0064) // APICOM
	WAVE_FORMAT_G722_ADPCM              = WaveFormatTag(0x0065) // APICOM
	WAVE_FORMAT_DSAT                    = WaveFormatTag(0x0066) // Microsoft Corporation
	WAVE_FORMAT_DSAT_DISPLAY            = WaveFormatTag(0x0067) // Microsoft Corporation
	WAVE_FORMAT_VOXWARE_BYTE_ALIGNED    = WaveFormatTag(0x0069) // Voxware Inc.
	WAVE_FORMAT_VOXWARE_AC8             = WaveFormatTag(0x0070) // Voxware Inc.
	WAVE_FORMAT_VOXWARE_AC10            = WaveFormatTag(0x0071) // Voxware Inc.
	WAVE_FORMAT_VOXWARE_AC16            = WaveFormatTag(0x0072) // Voxware Inc.
	WAVE_FORMAT_VOXWARE_AC20            = WaveFormatTag(0x0073) // Voxware Inc.
	WAVE_FORMAT_VOXWARE_RT24            = WaveFormatTag(0x0074) // Voxware Inc.
	WAVE_FORMAT_VOXWARE_RT29            = WaveFormatTag(0x0075) // Voxware Inc.
	WAVE_FORMAT_VOXWARE_RT29HW          = WaveFormatTag(0x0076) // Voxware Inc.
	WAVE_FORMAT_VOXWARE_VR12            = WaveFormatTag(0x0077) // Voxware Inc.
	WAVE_FORMAT_VOXWARE_VR18            = WaveFormatTag(0x0078) // Voxware Inc.
	WAVE_FORMAT_VOXWARE_TQ40            = WaveFormatTag(0x0079) // Voxware Inc.
	WAVE_FORMAT_SOFTSOUND               = WaveFormatTag(0x0080) // Softsound, Ltd.
	WAVE_FORMAT_VOXARE_TQ60             = WaveFormatTag(0x0081) // Voxware Inc.
	WAVE_FORMAT_MSRT24                  = WaveFormatTag(0x0082) // Microsoft Corporation
	WAVE_FORMAT_G729A                   = WaveFormatTag(0x0083) // AT&T Laboratories
	WAVE_FORMAT_MVI_MV12                = WaveFormatTag(0x0084) // Motion Pixels
	WAVE_FORMAT_DF_G726                 = WaveFormatTag(0x0085) // DataFusion Systems (Pty) (Ltd)
	WAVE_FORMAT_DF_GSM610               = WaveFormatTag(0x0086) // DataFusion Systems (Pty) (Ltd)
	WAVE_FORMAT_ONLIVE                  = WaveFormatTag(0x0089) // OnLive! Technologies, Inc.
	WAVE_FORMAT_SBC24                   = WaveFormatTag(0x0091) // Siemens Business Communications Systems
	WAVE_FORMAT_DOLBY_AC3_SPDIF         = WaveFormatTag(0x0092) // Sonic Foundry
	WAVE_FORMAT_ZYXEL_ADPCM             = WaveFormatTag(0x0097) // ZyXEL Communications, Inc.
	WAVE_FORMAT_PHILIPS_LPCBB           = WaveFormatTag(0x0098) // Philips Speech Processing
	WAVE_FORMAT_PACKED                  = WaveFormatTag(0x0099) // Studer Professional Audio AG
	WAVE_FORMAT_RHETOREX_ADPCM          = WaveFormatTag(0x0100) // Rhetorex, Inc.
	IBM_FORMAT_MULAW                    = WaveFormatTag(0x0101) // IBM mu-law format
	IBM_FORMAT_ALAW                     = WaveFormatTag(0x0102) // IBM a-law format
	IBM_FORMAT_ADPCM                    = WaveFormatTag(0x0103) // IBM AVC Adaptive Differential PCM format
	WAVE_FORMAT_VIVO_G723               = WaveFormatTag(0x0111) // Vivo Software
	WAVE_FORMAT_VIVO_SIREN              = WaveFormatTag(0x0112) // Vivo Software
	WAVE_FORMAT_DIGITAL_G723            = WaveFormatTag(0x0123) // Digital Equipment Corporation
	WAVE_FORMAT_CREATIVE_ADPCM          = WaveFormatTag(0x0200) // Creative Labs, Inc
	WAVE_FORMAT_CREATIVE_FASTSPEECH8    = WaveFormatTag(0x0202) // Creative Labs, Inc
	WAVE_FORMAT_CREATIVE_FASTSPEECH10   = WaveFormatTag(0x0203) // Creative Labs, Inc
	WAVE_FORMAT_QUARTERDECK             = WaveFormatTag(0x0220) // Quarterdeck Corporation
	WAVE_FORMAT_FM_TOWNS_SND            = WaveFormatTag(0x0300) // Fujitsu Corporation
	WAVE_FORMAT_BZV_DIGITAL             = WaveFormatTag(0x0400) // Brooktree Corporation
	WAVE_FORMAT_VME_VMPCM               = WaveFormatTag(0x0680) // AT&T Labs, Inc.
	WAVE_FORMAT_OLIGSM                  = WaveFormatTag(0x1000) // Ing C. Olivetti & C., S.p.A.
	WAVE_FORMAT_OLIADPCM                = WaveFormatTag(0x1001) // Ing C. Olivetti & C., S.p.A.
	WAVE_FORMAT_OLICELP                 = WaveFormatTag(0x1002) // Ing C. Olivetti & C., S.p.A.
	WAVE_FORMAT_OLISBC                  = WaveFormatTag(0x1003) // Ing C. Olivetti & C., S.p.A.
	WAVE_FORMAT_OLIOPR                  = WaveFormatTag(0x1004) // Ing C. Olivetti & C., S.p.A.
	WAVE_FORMAT_LH_CODEC                = WaveFormatTag(0x1100) // Lernout & Hauspie
	WAVE_FORMAT_NORRIS                  = WaveFormatTag(0x1400) // Norris Communications, Inc.
	WAVE_FORMAT_SOUNDSPACE_MUSICOMPRESS = WaveFormatTag(0x1500) // AT&T Labs, Inc.
	WAVE_FORMAT_DVM                     = WaveFormatTag(0x2000) // FAST Multimedia AG
	WAVE_FORMAT_INTERWAV_VSC112         = WaveFormatTag(0x7150) // ?????
	WAVE_FORMAT_EXTENSIBLE              = WaveFormatTag(0xFFFE) //
)

type WFESpeaker uint32

const (
	SPEAKER_FRONT_LEFT            = WFESpeaker(0x00000001)
	SPEAKER_FRONT_RIGHT           = WFESpeaker(0x00000002)
	SPEAKER_FRONT_CENTER          = WFESpeaker(0x00000004)
	SPEAKER_LOW_FREQUENCY         = WFESpeaker(0x00000008)
	SPEAKER_BACK_LEFT             = WFESpeaker(0x00000010)
	SPEAKER_BACK_RIGHT            = WFESpeaker(0x00000020)
	SPEAKER_FRONT_LEFT_OF_CENTER  = WFESpeaker(0x00000040)
	SPEAKER_FRONT_RIGHT_OF_CENTER = WFESpeaker(0x00000080)
	SPEAKER_BACK_CENTER           = WFESpeaker(0x00000100)
	SPEAKER_SIDE_LEFT             = WFESpeaker(0x00000200)
	SPEAKER_SIDE_RIGHT            = WFESpeaker(0x00000400)
	SPEAKER_TOP_CENTER            = WFESpeaker(0x00000800)
	SPEAKER_TOP_FRONT_LEFT        = WFESpeaker(0x00001000)
	SPEAKER_TOP_FRONT_CENTER      = WFESpeaker(0x00002000)
	SPEAKER_TOP_FRONT_RIGHT       = WFESpeaker(0x00004000)
	SPEAKER_TOP_BACK_LEFT         = WFESpeaker(0x00008000)
	SPEAKER_TOP_BACK_CENTER       = WFESpeaker(0x00010000)
	SPEAKER_TOP_BACK_RIGHT        = WFESpeaker(0x00020000)
)
