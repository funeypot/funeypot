FROM rockylinux:9

ENV TZ=Asia/Shanghai

RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

COPY bin/linux/piston /usr/local/bin/sshless

EXPOSE 22

CMD ["sshless", "-addr", ":22"]
