FROM debian:stable-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
      git \
      build-essential \
      autoconf \
      automake \
      texinfo \
      libacl1-dev \
      libattr1-dev \
      ca-certificates \
      autopoint \
      gettext \ 
      wget \
      bison \
      gperf \
      python3 \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /usr/src

RUN git clone https://git.savannah.gnu.org/git/coreutils.git

WORKDIR /usr/src/coreutils

RUN git fetch --tags --quiet && \
    git checkout v9.7

RUN ./bootstrap

RUN FORCE_UNSAFE_CONFIGURE=1 ./configure

RUN make -j $(nproc)

COPY test.sh /test.sh

RUN chmod +x /test.sh