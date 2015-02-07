FROM golang:cross

RUN apt-get update && apt-get install -y \
		libpulse-dev \
		--no-install-recommends \
	&& rm -rf /var/lib/apt/lists/*
