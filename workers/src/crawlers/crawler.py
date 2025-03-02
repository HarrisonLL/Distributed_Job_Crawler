from typing import List
import re

'''
crawler interface
'''

class Crawler():
    def __init__(self, job_type, location) -> None:
        self.job_type = job_type
        self.location = location

    def get_jobs(self) -> List:
        pass

    def get_job_details(self, url) -> dict:
        pass

    def get_job_id_by_url(self, url) -> str:
        pass

    @staticmethod
    def get_job_id_by_url(url) -> str:
        if url.endswith('/'): url = url[:-1]
        pattern = r"/jobs/(\d+)"
        match = re.search(pattern, url)
        if match:
            return match.group(1)
        return None