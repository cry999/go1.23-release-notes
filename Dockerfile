FROM alpine

WORKDIR /boringssl

RUN apk add --no-cache curl cmake make git

RUN git clone --branch=master --depth=1 "https://boringssl.googlesource.com/boringssl" .

RUN cmake -B build
RUN make -C build

CMD [ "./build/tool/bssl" ]
