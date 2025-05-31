import json
import requests
from typing import List
from crawlers.crawler import Crawler
from bs4 import BeautifulSoup
import html
import re


class microsoft(Crawler):
    def __init__(self, job_type, location) -> None:
        super().__init__(job_type, location)
        self.max_page = 10
        self.MICROSOFT_URL = "https://gcsservices.careers.microsoft.com/search/api/v1/search?l=en_us&pg={}&pgSz=200&o=Relevance&flt=true"

    def parse_job_page(self, page_num: int):
        response = requests.get(self.MICROSOFT_URL.format(page_num))
        if response.status_code != 200:
            return None
        return json.loads(response.text)
    
    @staticmethod
    def clean_microsoft_description(raw_html: str) -> str:
        soup = BeautifulSoup(raw_html, "html.parser")
        text = soup.get_text(separator=" ")
        text = html.unescape(text)
        invisible_chars = ["\u200b", "\u200c", "\u200d", "\u202f", "\xa0"]
        for ch in invisible_chars:
            text = text.replace(ch, " ")
        text = re.sub(r"\n\s*\n+", "\n\n", text)
        text = re.sub(r"[ \t]+", " ", text)
        text = text.strip()
        return text

    def get_jobs(self) -> List[dict]:
        jobs = []
        first_page_data = self.parse_job_page(1)
        if not first_page_data:
            raise Exception(f"Failed to fetch data from {self.MICROSOFT_URL}")
        try:
            total_jobs = first_page_data["operationResult"]["result"]["totalJobs"]
            jobs_per_page = len(first_page_data["operationResult"]["result"]["jobs"])
        except (KeyError, TypeError):
            raise Exception(f"Invalid response format. first_page_data: {first_page_data}")

        total_pages = min((total_jobs + jobs_per_page - 1) // jobs_per_page, self.max_page)

        for page in range(1, total_pages + 1):
            parsed_data = self.parse_job_page(page)
            if not parsed_data:
                continue
            job_entries = parsed_data.get("operationResult", {}).get("result", {}).get("jobs", [])
            for job in job_entries:
                location = job.get("properties", {}).get("primaryLocation", "")
                title = job.get("title", "")
                if (self.location.lower() == "usa" and "united states" not in location.lower()) or (
                    self.location.lower() != "usa" and self.location.lower() not in location.lower()):
                    continue
                if self.job_type.lower() not in title.lower():
                    continue
                description = job.get("properties", {}).get("description", "")
                if description:
                    description = microsoft.clean_microsoft_description(description)
                jobs.append({
                    "job_id": job.get("jobId", ""),
                    "url": f"https://jobs.careers.microsoft.com/global/en/job/{job.get("jobId", "")}",
                    "title": job.get("title", ""),
                    "date_posted": job.get("postingDate", ""),
                    "description": description,
                    "location": location,
                })
                print(description)
                print()
        return jobs
