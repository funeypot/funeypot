FROM rockylinux:9

ENV TZ=Asia/Shanghai

RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

COPY bin/linux/sshless /usr/local/bin/sshless

EXPOSE 2222

CMD ["sshless"]
