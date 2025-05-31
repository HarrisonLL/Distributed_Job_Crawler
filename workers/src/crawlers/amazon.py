from typing import List
from selenium import webdriver
from selenium.webdriver.common.by import By
from crawlers.crawler import Crawler
from bs4 import BeautifulSoup, Tag
import time
import requests


class amazon(Crawler):
    def __init__(self, job_type, location) -> None:
        super().__init__(job_type, location)
        self.AMAZONURL = "https://www.amazon.jobs/en/"
        self.max_page = 2 # max page per crawling
        self.init_web_driver(webdriver)

    def get_jobs(self) -> List:
        jobs = []
        for i in range(0, 10*self.max_page, 10):
            query = f"search?offset={i}&result_limit=10&sort=recent"
            query += f"&base_query={self.job_type}&country=USA&job_type%5B%5D=Full-Time&is_manager%5B%5D=0"
            if self.location != "USA":
                query += f"&state%5B%5D={self.location.capitalize()}"
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
    
    @staticmethod
    def _strip_text_after_two_br(html: str) -> str:
        soup = BeautifulSoup(html, 'html.parser')
        p_tag = soup.find('p')
        if not p_tag:
            return ""
        contents = p_tag.contents
        for i in range(len(contents) - 1):
            if isinstance(contents[i], Tag) and contents[i].name == 'br' and \
            isinstance(contents[i+1], Tag) and contents[i+1].name == 'br':
                new_contents = contents[:i]
                break
        else:
            new_contents = contents
        new_p = soup.new_tag('p')
        for item in new_contents:
            new_p.append(item)
        return new_p.get_text(separator='\n').strip()

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
            soup = BeautifulSoup(response.content, 'html.parser')
            job_title = soup.find('h1', class_='title')
            if job_title:
                job_details['title'] = job_title.text.strip()
            description_section = soup.find('h2', string='DESCRIPTION')
            if description_section:
                description = description_section.find_next('p')
                if description:
                    job_details['description'] = description.get_text(separator='\n').strip()
            for qualification in ["BASIC", "PREFERRED"]:
                qualification_section = soup.find('h2', string=f'{qualification} QUALIFICATIONS')
                if qualification_section:
                    qualification_content = qualification_section.find_next('p')
                    if qualification_content:
                        if 'qualifications' not in job_details:
                            job_details['qualifications'] = ''
                        qualification_content = amazon._strip_text_after_two_br(str(qualification_content))
                        job_details['qualifications'] += qualification + ': \n' + qualification_content + '\n'
        return job_details
