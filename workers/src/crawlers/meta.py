from typing import List
from seleniumwire import webdriver
from seleniumwire.utils import decode
from selenium.webdriver.chrome.service import Service
from selenium.webdriver.chrome.options import Options
from crawlers.crawler import Crawler
from bs4 import BeautifulSoup
from datetime import date
import requests
import json
import time
import os
import re


class meta(Crawler):
    def __init__(self, job_type, location) -> None:
        super().__init__(job_type, location)
        self.METAURL = "https://www.metacareers.com/jobs/"
        self.html_save_path = os.getenv('HTML_PATH', '/app/html_data')
    
    def _init_driver(self) -> None:
        chrome_driver_path = os.getenv('WEB_DRIVER_PATH', '/usr/local/bin/chromedriver')
        chrome_options = Options()
        chrome_options.add_argument('--headless')
        chrome_options.add_argument('--no-sandbox')
        chrome_options.add_argument('--disable-dev-shm-usage')
        service = Service(chrome_driver_path)
        driver = webdriver.Chrome(service=service, options=chrome_options)
        driver.maximize_window()
        self.driver = driver
   
    def _parse_job_page(self, url):
        self._init_driver()
        self.driver.get(url)
        time.sleep(10)
        for request in self.driver.requests:
            if request.response:
                if request.url == "https://www.metacareers.com/graphql":
                    body = decode(request.response.body, request.response.headers.get('Content-Encoding', 'identity'))
                    parsed = json.loads(body)["data"]
                    if "job_search" in parsed:
                        self.driver.quit()
                        return parsed
        self.driver.quit()
        return None
    
    def get_jobs(self) -> List:
        jobs = []
        query = f"?q={self.job_type.replace(' ', '%20')}"
        query += "&leadership_levels[0]=Individual%20Contributor&sort_by_new=true"
        query += "&offices[0]=New%20York%2C%20NY&offices[1]=Menlo%20Park%2C%20CA"
        parsed = self._parse_job_page(self.METAURL + query)["job_search"]
        for j in parsed:
            location = "N/A"
            if len(j["locations"]) > 0:
                location = j["locations"][0]
            jobs.append(
                {
                'title': j["title"],
                'desc': j["teams"][0],
                'location': location,
                'url': f"https://www.metacareers.com/jobs/{j['id']}"
                })
        return jobs

    def get_job_details(self, url) -> dict:
        response = requests.get(url)
        max_retry = 3
        i = 1
        while response.status_code != 200:
            i += 1
            response = requests.get(url)
            if i >= max_retry:
                break
        if response.status_code == 200:
            job_id = self.get_job_id_by_url(url)
            job_details = dict()
            self.save_job_details_to_html(response.text, self.html_save_path, f'meta_{job_id}.html')
            soup = BeautifulSoup(response.text, 'html.parser')

            if soup.find('title') is not None:
                job_details['job_title'] = soup.find('title').text
            
            if soup.find('script', type='application/ld+json') is not None:
                description_tag = soup.find('script', type='application/ld+json')
                description_json = json.loads(description_tag.string)
                job_details['description'] = description_json.get('description', '')
                job_details['responsibilities'] = description_json.get('responsibilities', '')
                job_details['qualifications'] = description_json.get('qualifications', '')
                job_details['locations'] = [location['address']['addressLocality'] + ", " + location['address']['addressRegion'] for location in description_json.get('jobLocation','')]
                job_details['employment_type'] = description_json.get('employmentType', '')
                job_details['date_posted'] = description_json.get('datePosted', '')
                job_details['valid_through'] = description_json.get('validThrough',)
            return job_details
        return None

    def get_job_id_by_url(self, url) -> str:
        pattern = r"/jobs/(\d+)"
        match = re.search(pattern, url)
        if match:
            return match.group(1)
        return None


if __name__ == '__main__':
    job_url = "https://www.metacareers.com/jobs/1341194059868794"
    job_url2 = "https://www.metacareers.com/jobs/774198984091403"
    crawler = meta('software engineer', 'USA')
    print(crawler.get_job_details(job_url))
