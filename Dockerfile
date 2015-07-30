FROM golang:cross

RUN apt-get clean && apt-get update && apt-get install -y \
		g++ \
		libpulse-dev \
		--no-install-recommends \
	&& rm -rf /var/lib/apt/lists/*
