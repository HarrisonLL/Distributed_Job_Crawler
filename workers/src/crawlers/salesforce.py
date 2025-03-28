import requests
import time
from typing import List
from bs4 import BeautifulSoup
from .crawler import Crawler

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
        if self.location.lower() in  self.region_map:
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
                title = title_tag.get_text(strip=True)
                if self.job_type.lower() not in title.lower():
                    continue
                location_tags = card.select("ul.locations li")
                if not title_tag:
                    continue
                path = title_tag["href"]
                url = "https://careers.salesforce.com" + path
                job_id = self.get_job_id_by_url(path, pattern=r"/jobs/(jr\d+)")
                
                locations = [li.get_text(strip=True) for li in location_tags]
                jobs.append({
                    "job_id": job_id,
                    "title": title,
                    "url": url,
                    "location": ", ".join(locations),
                })
            time.sleep(2)
        return jobs
