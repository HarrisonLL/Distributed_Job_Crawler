import datetime
import argparse
import os, json, requests
from crawlers import amazon, meta
from typing import List
from mongo_client import get_db, job_exists, save_job_url_to_db, save_job_details_to_db


def init_crawler(company: str, job_type: str, location: str):
    crawlers = {
        'amazon': amazon.amazon(job_type, location),
        'meta': meta.meta(job_type, location)
    }
    if company.lower() not in crawlers:
        raise ValueError('Current company not supported')
    return crawlers[company.lower()]

def _patch_data(data: dict, GS_URL:str, task_id:str) -> None:
    res = requests.patch(f'{GS_URL}/api/v1/tasks/{task_id}', data=json.dumps(data))
    print(res.json())

def _get_data(GS_URL:str, task_id:str) -> dict:
    res = requests.get(f'{GS_URL}/api/v1/tasks/{task_id}')
    if res.status_code != 200:
        print('Failed to get task details')
        return dict()
    else:
        return res.json()

def _crawl_individual_jobs(new_jobs:List[str], GS_URL:str, task_id:str, crawler, db) -> None:
    success = []
    failure = []
    for i,job in enumerate(new_jobs):
        print(f'Now scraping job: {job['url']}', flush=True)
        job_id = crawler.get_job_id_by_url(job['url'])
        save_job_url_to_db(db, job_id, job['url'])
        details = crawler.get_job_details(job['url'])
        if len(details.keys()) == 0:
            failure.append(job_id)
        else:
            details["crawled_datetime"] = datetime.datetime.now().strftime("%m/%d/%Y, %H:%M:%S")
            save_job_details_to_db(db, job_id, details)
            success.append(job_id)
        if GS_URL:
            data = {
                "completion_rate": (i + 1) / len(new_jobs),
                "failed_job_ids": failure,
                "success_job_ids": success,
            }
            _patch_data(data, GS_URL, task_id)
    if GS_URL:
        data = {"status": 4}
        _patch_data(data, GS_URL, task_id)


def retry_task(task_id: str, parent_task_id: str):
    GS_URL = os.getenv("GS_URL", None)
    task_details = _get_data(GS_URL, parent_task_id)
    failed_job_ids = task_details.get('failed_job_ids', [])
    prev_args = task_details.get('args', dict())
    if len(failed_job_ids) == 0 or len(prev_args) == 0:
        print('Nothing to retry')
        return
    company, job_type, location = prev_args['company'], prev_args['job_type'], prev_args['location']
    db = get_db(company)
    crawler = init_crawler(company, job_type, location)
    _crawl_individual_jobs(failed_job_ids, GS_URL, task_id, crawler, db)


def process_task(company: str, job_type: str, location: str, task_id: str):
    GS_URL = os.getenv("GS_URL", None)
    db = get_db(company)
    crawler = init_crawler(company, job_type, location)
    jobs = crawler.get_jobs()
    print(f'{len(jobs)} Jobs Found', flush=True)
    if len(jobs) == 0:
        if GS_URL:
            data = {"status": 3}
            _patch_data(data, GS_URL, task_id)
        return
    if GS_URL:
        data = {"status": 2}
        _patch_data(data, GS_URL, task_id)

    new_jobs = []
    for job in jobs:
        job_id = crawler.get_job_id_by_url(job['url'])
        if not job_exists(db, job_id):
            new_jobs.append(job)
    _crawl_individual_jobs(new_jobs, GS_URL, task_id, crawler, db)


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description='Job Crawler Script')
    parser.add_argument('--job_type', type=str, help='Job type for the job search')
    parser.add_argument('--location', type=str, help='Location for the job search')
    parser.add_argument('--company', type=str, help='Company for the job search')
    parser.add_argument('--task_id', type=str, help='Current task id')
    parser.add_argument('--retry', type=str, help='If the task is retried task')
    parser.add_argument('--parent_task_id', type=str, help='Parent task id')
    args = parser.parse_args()
    if args.retry and args.retry.lower() == "true":
        retry_task(args.task_id,args.parent_task_id)
    else:
        process_task(args.company, args.job_type, args.location, args.task_id)
