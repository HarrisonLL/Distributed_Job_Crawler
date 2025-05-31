import time
import requests
from typing import List
from crawlers.crawler import Crawler
from bs4 import BeautifulSoup


class linkedInPosts(Crawler):
    def __init__(self, job_type, location):
        super().__init__(job_type, location)
        self.total_pages = 10
        self.location_map = {
            "new york": "New York Metropolitan Area",
            "washington": "Seattle Metropolitan Area",
            "california": "San Francisco Bay Area"
        }
        self.headers = {
            "User-Agent": "Mozilla/5.0"
        }
    
    def get_jobs(self) -> List:
        jobs = []
        for page in range( self.total_pages):
            start = page * 25
            url = (
                "https://www.linkedin.com/jobs-guest/jobs/api/seeMoreJobPostings/search"
                f"?keywords={self.job_type}&location={self.location_map[self.location.lower()]}&start={start}"
            )
            response = requests.get(url, headers=self.headers)
            if response.status_code != 200:
                raise Exception(f"Failed to fetch page {page + 1}: Status code {response.status_code}")

            soup = BeautifulSoup(response.text, 'html.parser')
            job_cards = soup.find_all('li')
            if not job_cards:
                raise Exception(("No more job postings found."))
            for job in job_cards:
                hiring_flag = job.find('span', class_='job-posting-benefits__text')
                if not (hiring_flag and 'Actively Hiring' in hiring_flag.text):
                    continue
                title_elem = job.find('h3', class_='base-search-card__title')
                if title_elem and "intern" in title_elem.get_text(strip=True).lower():
                    continue
                company_elem = job.find('h4', class_='base-search-card__subtitle')
                link_elem = job.find('a', class_='base-card__full-link', href=True)
                time_elem = job.find('time', class_='job-search-card__listdate')
                job_data = {
                    'title': title_elem.get_text(strip=True) if title_elem else None,
                    'company': company_elem.get_text(strip=True) if company_elem else '',
                    'link': link_elem['href'].strip() if link_elem else None,
                    'posting_date': time_elem['datetime'] if time_elem and time_elem.has_attr('datetime') else '',
                }
                jobs.append(job_data)
            time.sleep(5)
        return jobs