from typing import List

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

    def save_job_details_to_html(self, html_content:str, save_path:str, html_name:str):
        with open(f'{save_path}/{html_name}', 'w') as f:
            f.write(html_content)
    


