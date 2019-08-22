FROM scratch
VOLUME /
ADD config.yml /
ADD goFileBed-linux /
CMD ["/goFileBed-linux"]