FROM fedora:34

RUN sed -i 's/enabled=1/enabled=0/' /etc/yum.repos.d/fedora-cisco-openh264.repo \
	&& sed -i 's/enabled=1/enabled=0/' /etc/yum.repos.d/fedora-modular.repo \
	&& sed -i 's/enabled=1/enabled=0/' /etc/yum.repos.d/fedora-updates-modular.repo

RUN dnf -y update \
	&& dnf -y install pip \
	&& dnf clean all

RUN pip install --upgrade pip && pip install pycodestyle

WORKDIR /src
COPY ./ ./

RUN pip install --requirement requirements.txt
