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
      strace \
      procps \
      sudo \
      libexpect-perl \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /usr/src

RUN useradd -m user

RUN echo "user ALL=(ALL) NOPASSWD: ALL" >> /etc/sudoers

RUN chown -R user:user /usr/src/

COPY test.sh /test.sh

RUN chmod +x /test.sh

USER user

RUN git clone https://github.com/coreutils/coreutils.git

WORKDIR /usr/src/coreutils

RUN git fetch --tags --quiet && \
    git checkout v9.7

RUN git config submodule.gnulib.url https://github.com/coreutils/gnulib.git

RUN ./bootstrap

RUN ./configure

RUN make -j $(nproc)

