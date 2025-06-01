import json
import requests
from typing import List, Optional
from crawlers.crawler import Crawler
from bs4 import BeautifulSoup


class google(Crawler):
    BASE_URL = "https://www.google.com/about/careers/applications/jobs/results"
    API_URL = "https://www.google.com/about/careers/applications/_/HiringCportalFrontendUi/data/batchexecute"
    HEADERS = {"Content-Type": "application/x-www-form-urlencoded"}
    MAX_PAGES = 20

    def __init__(self, job_type: str, location: str):
        super().__init__(job_type, location)

    def _parse_job_page(self, page_num: int = 1) -> Optional[List]:
        try:
            payload = {
                "f.req": f'[[["r06xKb","[[null,[],[],[],null,null,[],{page_num},null,null,null,null,null,[],[],null,[2]]]"]]]'
            }
            response = requests.post(self.API_URL, data=payload, headers=self.HEADERS)
            text = "\n".join(response.text.splitlines()[2:])
            parsed = json.loads(json.loads(text)[0][2])
            return parsed[0] if parsed else []
        except Exception as e:
            raise Exception(f"Failed to parse page {page_num}: {e}")

    @staticmethod
    def _clean_html(html: str) -> str:
        if html:
            soup = BeautifulSoup(html, 'html.parser')
            return soup.get_text(separator='\n').strip()
        return ""

    def _extract_html_field(self, field) -> str:
        return self._clean_html(field[1] if isinstance(field, list) and len(field) > 1 else "")

    def get_jobs(self) -> List[dict]:
        jobs = []
        for i in range(self.MAX_PAGES):
            job_entries = self._parse_job_page(i)
            for entry in job_entries:
                try:
                    title = entry[1]
                    location = entry[9][0][0] if entry[9] and entry[9][0] else ""
                    if self.job_type.lower() not in title.lower() or self.location.lower() not in location.lower():
                        continue
                    job = {
                        "title": title,
                        "job_id": entry[0],
                        "location": location,
                        "url": f"{self.BASE_URL}/{entry[0]}",
                        "description": self._extract_html_field(entry[3]),
                        "summary": self._extract_html_field(entry[10]),
                        "qualifications": self._extract_html_field(entry[4]),
                    }
                    jobs.append(job)
                except Exception as e:
                    print(f"Error processing job entry: {e}", flush=True)
        return jobs
