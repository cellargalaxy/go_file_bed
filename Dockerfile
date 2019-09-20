FROM scratch
ENV TOKEN token
ENV LISTEN_ADDRESS 0.0.0.0:8880
ENV FILE_BED_PATH file_bed
VOLUME /file_bed
ADD goFileBed /
CMD ["/goFileBed"]