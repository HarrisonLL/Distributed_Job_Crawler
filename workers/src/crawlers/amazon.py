from typing import List
from selenium import webdriver
from selenium.webdriver.chrome.service import Service
from selenium.webdriver.common.by import By
from selenium.webdriver.chrome.options import Options
from crawlers.crawler import Crawler
from bs4 import BeautifulSoup
import time
import requests
import os
import re


class amazon(Crawler):
    def __init__(self, job_type, location) -> None:
        super().__init__(job_type, location)
        self.AMAZONURL = "https://www.amazon.jobs/en/"
        self.web_driver_path = os.getenv('WEB_DRIVER_PATH', '/usr/local/bin/chromedriver')
        self.max_page = 2 # max page per crawling
        self.html_save_path = os.getenv('HTML_PATH', '/app/html_data')

    def _init_driver(self) -> None:
        chrome_options = Options()
        chrome_options.add_argument("--headless")
        chrome_options.add_argument("--no-sandbox")
        chrome_options.add_argument("--disable-dev-shm-usage")
        service = Service(self.web_driver_path)
        self.driver = webdriver.Chrome(service=service, options=chrome_options)
    
    def get_jobs(self) -> List:
        jobs = []
        self._init_driver()
        for i in range(0, 10*self.max_page, 10):
            query = f"search?offset={i}&result_limit=10&sort=recent"
            query += f"&base_query={self.job_type}&country={self.location}"
            url = self.AMAZONURL + query
            self.driver.get(url)
            time.sleep(5)
            job_listings = self.driver.find_elements(By.CLASS_NAME, "job-tile")
            for job in job_listings:
                title = job.find_element(By.CLASS_NAME, "job-title").text
                location = job.find_element(By.CLASS_NAME, "location-and-id").text
                url = job.find_element(By.TAG_NAME, "a").get_attribute("href")
                jobs.append({
                    'title': title,
                    'location': location,
                    'url': url
                })
        self.driver.quit()
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
        job_details = dict()
        if response.status_code == 200:
            job_id = self.get_job_id_by_url(url)
            self.save_job_details_to_html(response.text, self.html_save_path, f'amazon_{job_id}.html')
            soup = BeautifulSoup(response.content, 'html.parser')
            job_title = soup.find('h1', class_='title')
            if job_title:
                job_details['title'] = job_title.text.strip()
            description_section = soup.find('h2', string='DESCRIPTION')
            if description_section:
                description = description_section.find_next('p')
                if description:
                    job_details['description'] = description.get_text(separator='\n').strip()
            basic_qualifications_heading = soup.find('h2', string='BASIC QUALIFICATIONS')
            if basic_qualifications_heading:
                basic_qualifications_section = basic_qualifications_heading.find_next('p')
                if basic_qualifications_section:
                    job_details['basic_qualifications'] = basic_qualifications_section.get_text(separator='\n').strip()
            preferred_qualifications_heading = soup.find('h2', string='PREFERRED QUALIFICATIONS')
            if preferred_qualifications_heading:
                preferred_qualifications_section = preferred_qualifications_heading.find_next('p')
                if preferred_qualifications_section:
                    job_details['preferred_qualifications'] = preferred_qualifications_section.get_text(separator='\n').strip()
        return job_details

    def get_job_id_by_url(self, url) -> str:
        pattern = r"/jobs/(\d+)/"
        match = re.search(pattern, url)
        if match:
            return match.group(1)
        return None
