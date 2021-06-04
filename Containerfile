FROM fedora:34

RUN dnf -y update \
	&& dnf -y install pip \
	&& dnf clean all

RUN pip install --upgrade pip && pip install pycodestyle

WORKDIR /src
COPY ./ ./

RUN pip install --requirement requirements.txt
