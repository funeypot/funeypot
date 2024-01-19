FROM rockylinux:9

ENV TZ=Asia/Shanghai

RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

COPY bin/linux/funeypot /usr/local/bin/funeypot

EXPOSE 2222

ENTRYPOINT ["funeypot"]
