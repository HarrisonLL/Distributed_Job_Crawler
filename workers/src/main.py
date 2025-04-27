import datetime
import argparse
import os, json, requests, logging
from crawlers import amazon, meta, google, uber, salesforce
from typing import List
from mongo_client import get_db, job_exists, save_job_url_to_db, save_job_details_to_db, save_job_type_to_db, get_job_type

logging.basicConfig(format="[%(asctime)s] [%(levelname)s] - %(message)s")
logger = logging.getLogger()

def init_crawler(company: str, job_type: str, location: str):
    crawlers = {
        'amazon': amazon.amazon(job_type, location),
        'meta': meta.meta(job_type, location),
        'google': google.google(job_type, location),
        'uber': uber.uber(job_type, location),
        'salesforce': salesforce.salesforce(job_type, location)
    }
    if company.lower() not in crawlers:
        raise ValueError('Current company not supported')
    return crawlers[company.lower()]

def _patch_data(data: dict, GS_URL:str, task_id:str) -> None:
    if not GS_URL: return
    res = requests.patch(f'{GS_URL}/api/v1/tasks/{task_id}', data=json.dumps(data))
    logger.info(f'PATCH Task {task_id} {res.status_code} {res.json()}')
    if res.status_code != 200:
        logger.warning(f'PATCH Task {task_id} {res.status_code} {res.json()}')

def _crawl_individual_jobs(new_jobs:List[dict], stored_job_ids: List[str], GS_URL:str, task_id:str, job_type:str, crawler, db) -> None:
    success = []
    for job in new_jobs:
        try:
            details = crawler.get_job_details(job['url'])
        except Exception as e:
            logger.error(e, exc_info=True)
            continue
        if details is None or len(details.keys()) == 0:
            continue
        else: # only when success then save entry
            details["crawled_datetime"] = datetime.datetime.now().strftime("%m/%d/%Y, %H:%M:%S")
            details["job_types"] = [job_type]
            try:
                save_job_url_to_db(db, job["job_id"], job["url"])
                save_job_details_to_db(db, job["job_id"], details)
                success.append(job["job_id"])
            except Exception as e:
                logger.error(e, exc_info=True)
    success += stored_job_ids
    data = {
        "completion_rate": len(success) / len(new_jobs),
        "success_job_ids": success,
        "status": 4
    }
    print(success, flush=True)
    _patch_data(data, GS_URL, task_id)

def process_task(company: str, job_type: str, location: str, task_id: str):
    GS_URL = os.getenv("GS_URL", "")
    try:
        _patch_data({"status": 2}, GS_URL, task_id)
        db = get_db(company)
        crawler = init_crawler(company, job_type, location)
        jobs = crawler.get_jobs()
    except Exception as e:
        logger.error(e, exc_info=True)
        _patch_data({"status": 3}, GS_URL, task_id)
        return
    # Check if found jobs are already in DB and perform duplication check
    if len(jobs) == 0:
        logger.info("No new jobs are found.")
        _patch_data({"status": 4}, GS_URL, task_id)
        return
    
    new_jobs = []
    stored_job_ids = []
    for job in jobs:
        if 'job_id' not in job:
            job['job_id'] = crawler.get_job_id_by_url(job['url'])
        if not job_exists(db, job['job_id']):
            new_jobs.append(job)
        else:
            stored_job_types = get_job_type(db, job['job_id'])
            if job_type not in stored_job_types:
                stored_job_types.append(job_type)
                current_time = datetime.datetime.now().strftime("%m/%d/%Y, %H:%M:%S")
                save_job_type_to_db(db, job['job_id'], stored_job_types, current_time)
                stored_job_ids.append(job['job_id'])
    if len(new_jobs) == 0 and len(stored_job_ids) == 0:
        logger.info("No new jobs are found.")
        _patch_data({"status": 4}, GS_URL, task_id)
        return

    print(new_jobs, stored_job_ids, flush=True)
    # Parse webpage and save new job to DB
    if company in ['google', 'uber']: 
        # companies that skip parsing step
        for job in new_jobs:
            job["crawled_datetime"] = datetime.datetime.now().strftime("%m/%d/%Y, %H:%M:%S")
            job["job_types"] = [job_type]
            save_job_url_to_db(db, job["job_id"], job["url"])
            save_job_details_to_db(db, job["job_id"], job)
        ret_job_ids = [job["job_id"] for job in new_jobs] + stored_job_ids
        data = { "completion_rate": 1,  "success_job_ids": ret_job_ids, "status": 4}
        _patch_data(data, GS_URL, task_id)
    else: 
        # companies that need parsing step
        if len(new_jobs) > 0:
            _crawl_individual_jobs(new_jobs, stored_job_ids, GS_URL, task_id, job_type, crawler, db)
        else:
            data = { "completion_rate": 1,  "success_job_ids": stored_job_ids, "status": 4}
            _patch_data(data, GS_URL, task_id)


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description='Job Crawler Script')
    parser.add_argument('--job_type', type=str, help='Job type for the job search')
    parser.add_argument('--location', type=str, help='Location for the job search')
    parser.add_argument('--company', type=str, help='Company for the job search')
    parser.add_argument('--task_id', type=str, help='Current task id')

    args = parser.parse_args()
    process_task(args.company, args.job_type, args.location, args.task_id)
