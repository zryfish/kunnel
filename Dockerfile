
ARG GOLANG_IMAGE=golang:1.16.5
FROM ${GOLANG_IMAGE} as binary_tools_context

ENV OUTDIR=/out
RUN mkdir -p ${OUTDIR}/usr/local/bin/

WORKDIR /workspace
ADD . /workspace/

RUN make all
RUN mv /workspace/bin/{server,client,kubectl-kn} ${OUTDIR}/usr/local/bin/

##############
# Final image
#############

FROM alpine:3.11 

COPY --from=base_os_context /out/ /
COPY --from=build_context /out/ /

WORKDIR /
CMD ["sh"]
