# FROM ubuntu:22.04
# ARG DEBIAN_FRONTEND=noninteractive
# ENV TZ=America/New_York
# WORKDIR /app

# RUN apt update && apt install -y curl unzip gnupg2 wget \
#     && apt install -y python3.10 python3.10-dev python3.10-distutils python3.10-tk python3-pip

# # Install Chrome WebDriver
# RUN CHROMEDRIVER_VERSION=`curl -sS chromedriver.storage.googleapis.com/LATEST_RELEASE` && \
#     mkdir -p /opt/chromedriver-$CHROMEDRIVER_VERSION && \
#     curl -sS -o /tmp/chromedriver_linux64.zip http://chromedriver.storage.googleapis.com/$CHROMEDRIVER_VERSION/chromedriver_linux64.zip && \
#     unzip -qq /tmp/chromedriver_linux64.zip -d /opt/chromedriver-$CHROMEDRIVER_VERSION && \
#     rm /tmp/chromedriver_linux64.zip && \
#     chmod +x /opt/chromedriver-$CHROMEDRIVER_VERSION/chromedriver && \
#     ln -fs /opt/chromedriver-$CHROMEDRIVER_VERSION/chromedriver /usr/local/bin/chromedriver

# # Install Google Chrome
# RUN wget https://dl.google.com/linux/direct/google-chrome-stable_current_arm64.deb \
#     && dpkg -i google-chrome-stable_current_arm64.deb \
#     && apt --fix-broken install
# # RUN mkdir -p /etc/apt/keyrings && \
# #     curl -sS https://dl.google.com/linux/linux_signing_key.pub | gpg --dearmor -o /etc/apt/keyrings/google-chrome.gpg && \
# #     echo "deb [arch=amd64 signed-by=/etc/apt/keyrings/google-chrome.gpg] http://dl.google.com/linux/chrome/deb/ stable main" | tee /etc/apt/sources.list.d/google-chrome.list > /dev/null && \
# #     apt-get update && \
# #     apt-get install -y google-chrome-stable && \
# #     rm -rf /var/lib/apt/lists/*

# COPY requirements.txt /app/

# RUN python3.10 -m pip install --no-cache-dir --ignore-installed -r requirements.txt && rm requirements.txt

# RUN update-alternatives --install /usr/bin/python python /usr/bin/python3.10 1 \
#     && update-alternatives --set python /usr/bin/python3.10

# ENTRYPOINT ["python", "main.py"]


FROM --platform=linux/amd64 python:3.12.4

RUN apt-get update && apt-get install -y curl unzip gnupg2 \
    && rm -rf /var/lib/apt/lists/*

# Install Chrome WebDriver
RUN CHROMEDRIVER_VERSION=`curl -sS chromedriver.storage.googleapis.com/LATEST_RELEASE` && \
    mkdir -p /opt/chromedriver-$CHROMEDRIVER_VERSION && \
    curl -sS -o /tmp/chromedriver_linux64.zip http://chromedriver.storage.googleapis.com/$CHROMEDRIVER_VERSION/chromedriver_linux64.zip && \
    unzip -qq /tmp/chromedriver_linux64.zip -d /opt/chromedriver-$CHROMEDRIVER_VERSION && \
    rm /tmp/chromedriver_linux64.zip && \
    chmod +x /opt/chromedriver-$CHROMEDRIVER_VERSION/chromedriver && \
    ln -fs /opt/chromedriver-$CHROMEDRIVER_VERSION/chromedriver /usr/local/bin/chromedriver

# Install Google Chrome
# RUN curl -sS -o - https://dl-ssl.google.com/linux/linux_signing_key.pub | apt-key add - && \
#     echo "deb http://dl.google.com/linux/chrome/deb/ stable main" >> /etc/apt/sources.list.d/google-chrome.list && \
#     apt-get -yqq update && \
#     apt-get -yqq install google-chrome-stable && \
#     rm -rf /var/lib/apt/lists/*

RUN curl -sS -o - https://dl-ssl.google.com/linux/linux_signing_key.pub | apt-key add - && \
    echo "deb [arch=amd64] http://dl.google.com/linux/chrome/deb/ stable main" >> /etc/apt/sources.list.d/google-chrome.list && \
    apt-get -yqq update && \
    apt-get -yqq install google-chrome-stable=114.0.5735.90-1 && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY requirements.txt /app/

RUN pip install --no-cache-dir -r requirements.txt && rm requirements.txt

COPY src /app

ENTRYPOINT ["python", "main.py"]
