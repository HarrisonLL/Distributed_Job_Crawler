from typing import List
import re, os
from selenium.webdriver.chrome.service import Service
from selenium.webdriver.chrome.options import Options

'''
crawler interface
'''

class Crawler():
    def __init__(self, job_type, location) -> None:
        self.job_type = job_type
        self.location = location
        self.driver = None # default to none, call init_web_driver to init
    
    def init_web_driver(self, webdriver):
        chrome_driver_path = os.getenv('WEB_DRIVER_PATH', '/usr/local/bin/chromedriver')
        chrome_options = Options()
        chrome_options.add_argument('--headless')
        chrome_options.add_argument('--no-sandbox')
        chrome_options.add_argument('--disable-dev-shm-usage')
        service = Service(chrome_driver_path)
        self.driver = webdriver.Chrome(service=service, options=chrome_options)
        self.driver.maximize_window()

    def get_jobs(self) -> List:
        pass

    def get_job_details(self, url) -> dict:
        pass

    @staticmethod
    def get_job_id_by_url(url, pattern = r"/jobs/(\d+)") -> str:
        if url.endswith('/'): url = url[:-1]
        match = re.search(pattern, url)
        if match:
            return match.group(1)
        return None