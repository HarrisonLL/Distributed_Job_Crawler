import requests
from typing import List
from .crawler import Crawler

class uber(Crawler):
    def __init__(self, job_type, location):
        super().__init__(job_type, location)
        self.api_url = "https://www.uber.com/api/loadSearchJobsResults?localeCode=en"
        self.headers = {
            "Content-Type": "application/json",
            "Accept": "*/*",
            "Origin": "https://www.uber.com",
            "Referer": "https://www.uber.com/us/en/careers/list/",
            "User-Agent": (
                "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) "
                "AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.0.0 Safari/537.36"
            ),
            "x-csrf-token": "x",
            "x-uber-sites-page-edge-cache-enabled": "true",
        }

    def get_jobs(self) -> List[dict]:
        city_filters = []
        if self.location.lower() == "san francisco":
            city_filters.append({"country": "USA", "region": "California", "city": "San Francisco"})
        elif self.location.lower() == "seattle":
            city_filters.append({"country": "USA", "region": "Washington", "city": "Seattle"})
        elif self.location.lower() == "new york":
            city_filters.append({"country": "USA", "region": "New York", "city": "New York"})
        else:
            city_filters.extend([{"country": "USA", "region": "California", "city": "San Francisco"},
                {"country": "USA", "region": "Washington", "city": "Seattle"}, 
                {"country": "USA", "region": "New York", "city": "New York"}])

        payload = {
            "limit": 50,
            "page": 0,
            "params": {
                "department": [],
                "lineOfBusinessName": [],
                "location": city_filters,
                "programAndPlatform": [],
                "team": [],
                "query": self.job_type
            }
        }
        response = requests.post(self.api_url, json=payload, headers=self.headers)
        jobs = []
        if response.ok:
            data = response.json()
            results = data.get("data", {}).get("results", [])
            for job in results:
                job_id = str(job["id"])
                jobs.append({
                    "job_id": job_id,
                    "title": job["title"],
                    "location": f"{job['location']['city']}, {job['location']['region']}",
                    "url": f"https://www.uber.com/global/en/careers/list/{job_id}/",
                    "department": job.get("department", ""),
                    "team": job.get("team", ""),
                    "created_at": job.get("creationDate", "")
                })
        else:
            raise Exception(f"Uber API request failed: {response.status_code} \n {response.text}")
        return jobs
