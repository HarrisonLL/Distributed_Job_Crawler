import requests
import time
from typing import List
from bs4 import BeautifulSoup
from .crawler import Crawler
import re

class salesforce(Crawler):
    def __init__(self, job_type, location):
        super().__init__(job_type, location)
        self.base_url = "https://careers.salesforce.com/en/jobs/"
        self.region_map = {
            "new york": "New+York",
            "california": "California",
            "washington": "Washington"
        }
        self.max_page = 5
        self.request_header = {
                "accept": "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
                "accept-language": "en-US,en;q=0.9,zh-CN;q=0.8,zh;q=0.7",
                "upgrade-insecure-requests": "1",
                "User-Agent": (
                    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) "
                    "AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.0.0 Safari/537.36"
                    ),
                "sec-fetch-site": "same-origin",
                "sec-fetch-mode": "navigate",
                "sec-fetch-user": "?1",
                "sec-fetch-dest": "document",
                "sec-ch-ua": '"Chromium";v="134", "Not:A-Brand";v="24", "Google Chrome";v="134"',
                "sec-ch-ua-mobile": "?0",
                "sec-ch-ua-platform": '"macOS"',
        }

    def get_region_query_string(self):
        if not self.location:
            return ""
        region_params = []
        if self.location.lower() in self.region_map:
            region_params.append(f"&region={self.region_map[self.location.lower()]}")
        return "".join(region_params)

    def get_jobs(self) -> List[dict]:
        jobs = []
        for page in range(1, self.max_page+1):
            region_part = self.get_region_query_string()
            full_url = (
                self.base_url +
                f"?page={page}&search={self.job_type.replace(' ', '+')}&country=United+States+of+America" +
                "&type=Full+time&jobtype=Regular&pagesize=20"+
                region_part
            )
            self.request_header["referer"] = full_url
            response = requests.get(full_url, headers=self.request_header)
            if not response.ok:
                raise Exception(f"Salesforce job page request failed: {response.status_code}\n{response.text}")
            soup = BeautifulSoup(response.text, "html.parser")
            job_cards = soup.select("div.card.card-job")
            for card in job_cards:
                title_tag = card.select_one("h3.card-title a")
                if not title_tag:
                    continue
                title = title_tag.get_text(strip=True)
                if self.job_type.lower() not in title.lower():
                    continue
                path = title_tag["href"]
                url = "https://careers.salesforce.com" + path
                job_id = self.get_job_id_by_url(path, pattern=r"/jobs/(jr\d+)")
                jobs.append({
                    "job_id": job_id,
                    "title": title,
                    "url": url,
                })
            time.sleep(2)
        return jobs

    def get_job_details(self, url) -> dict:
        resp = requests.get(url)
        soup = BeautifulSoup(resp.text, 'html.parser')
        title = soup.find('h1', class_='hero-heading')
        title = title.text.strip() if title else None
        location_block = soup.select_one('.job-meta .multi-locations-list')
        locations = [li.text.strip() for li in location_block.find_all('li')] if location_block else []
        posting_time_tag = soup.find('time')
        posting_date = posting_time_tag['datetime'] if posting_time_tag else None
        job_id_tag = soup.find(string=re.compile(r'JR\d+'))
        job_id = job_id_tag.strip() if job_id_tag else None
        salary_tags = soup.select('.job-meta li')
        salaries = [tag.text.strip() for tag in salary_tags if 'Salary' in tag.text]
        description_tag = soup.select_one('div.job-detail article.cms-content')
        description = description_tag.get_text(separator='\n').strip() if description_tag else None
        apply_link_tag = soup.select_one('.js-apply-now')
        apply_url = apply_link_tag['href'] if apply_link_tag else None
        return {
            'title': title,
            'locations': locations,
            'posting_date': posting_date,
            'job_id': job_id,
            'salaries': salaries,
            'description': description,
            'apply_url': apply_url,
        }
