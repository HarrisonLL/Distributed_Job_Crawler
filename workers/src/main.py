import datetime
import argparse
import os, time, json, requests, logging
from crawlers import amazon, meta
from typing import List
from mongo_client import get_db, job_exists, save_job_url_to_db, save_job_details_to_db

logging.basicConfig(format="[%(asctime)s] [%(levelname)s] - %(message)s")
logger = logging.getLogger()

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
    logger.info(f'PATCH Task {task_id} {res.status_code} {res.json()}')

def _crawl_individual_jobs(new_jobs:List[str], GS_URL:str, task_id:str, crawler, db) -> None:
    success = []
    for i,job in enumerate(new_jobs):
        logger.info(f'Now scraping job: {job['url']}')
        try:
            job_id = crawler.get_job_id_by_url(job['url'])
            details = crawler.get_job_details(job['url'])
        except Exception as e:
            logger.error(e)
            continue
        if details is None or len(details.keys()) == 0:
            continue
        else: # only when success then save entry
            details["crawled_datetime"] = datetime.datetime.now().strftime("%m/%d/%Y, %H:%M:%S")
            try:
                save_job_url_to_db(db, job_id, job['url'])
                save_job_details_to_db(db, job_id, details)
                success.append(job_id)
            except Exception as e:
                logger.error(e)
                
        if GS_URL:
            data = {
                "completion_rate": (i + 1) / len(new_jobs),
                "success_job_ids": success,
            }
            _patch_data(data, GS_URL, task_id)
    if GS_URL:   _patch_data({"status": 4}, GS_URL, task_id)

def process_task(company: str, job_type: str, location: str, task_id: str):
    GS_URL = os.getenv("GS_URL", None)
    try:
        db = get_db(company)
        crawler = init_crawler(company, job_type, location)
        jobs = crawler.get_jobs()
    except Exception as e:
        logger.error(e)
        if GS_URL: _patch_data({"status": 3}, GS_URL, task_id)
        return
    if len(jobs) == 0:
        logger.info('No new jobs are found.')
        if GS_URL: _patch_data({"status": 4}, GS_URL, task_id)
        return
    if GS_URL: _patch_data({"status": 2}, GS_URL, task_id)

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

    args = parser.parse_args()
    process_task(args.company, args.job_type, args.location, args.task_id)
