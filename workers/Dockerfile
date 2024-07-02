FROM --platform=linux/amd64 python:3.12.4

RUN apt-get update && apt-get install -y curl unzip gnupg2 \
    && rm -rf /var/lib/apt/lists/*

RUN wget -q -O - https://dl-ssl.google.com/linux/linux_signing_key.pub | apt-key add - && \
    echo 'deb [arch=amd64] http://dl.google.com/linux/chrome/deb/ stable main' | tee /etc/apt/sources.list.d/google-chrome.list && \
    apt-get update -y && apt-get install google-chrome-stable -y

WORKDIR /app

RUN mkdir -p /app/html_data

COPY requirements.txt /app/

RUN pip install --no-cache-dir -r requirements.txt && rm requirements.txt

COPY install_driver.py /app/

RUN python /app/install_driver.py

ENV WEB_DRIVER_PATH=/root/.wdm/drivers/chromedriver/linux64/126.0.6478.126/chromedriver-linux64/chromedriver

COPY src /app

ENTRYPOINT ["python", "main.py"]
