from typing import List
from seleniumwire import webdriver
from seleniumwire.utils import decode
from crawlers.crawler import Crawler
from bs4 import BeautifulSoup
import requests
import json
import time


class meta(Crawler):
    def __init__(self, job_type, location) -> None:
        super().__init__(job_type, location)
        self.METAURL = "https://www.metacareers.com/jobs/"
        self.init_web_driver(webdriver)
   
    def _parse_job_page(self, url):
        self.driver.get(url)
        time.sleep(10)
        all_data = []
        for request in self.driver.requests:
            if request.response:
                if request.url == "https://www.metacareers.com/graphql":
                    body = decode(request.response.body, request.response.headers.get('Content-Encoding', 'identity'))
                    data = json.loads(body)["data"]
                    all_data.append(data)
                    if "job_search_with_featured_jobs" in data:
                        self.driver.quit()
                        try:
                            all_jobs =  data["job_search_with_featured_jobs"]["all_jobs"]
                        except KeyError:
                            raise Exception(f"Parsed data schema changed. \n {data["job_search_with_featured_jobs"].keys()}")
                        return all_jobs
        
        self.driver.quit()
        if len(all_data) > 0:
            raise Exception(f"Responsed data schema changed. \n {all_data}")
        raise Exception(f"Requests failed")
    
    def get_jobs(self) -> List:
        jobs = []
        query = f"?q={self.job_type.replace(' ', '%20')}"
        query += "&leadership_levels[0]=Individual%20Contributor&sort_by_new=true&roles[0]=Full%20time%20employment"
        query += "&offices[0]=Menlo%20Park%2C%20CA&offices[1]=Seattle%2C%20WA&offices[2]=New%20York%2C%20NY"
        parsed = self._parse_job_page(self.METAURL + query)
        for j in parsed:
            location = "N/A"
            if len(j["locations"]) > 0:
                location = j["locations"][0]
            jobs.append(
                {
                'title': j["title"],
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
        job_details = dict()
        if response.status_code == 200:
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
