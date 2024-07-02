import datetime
import argparse
from crawlers import amazon, meta
from mongo_client import get_db, job_exists, save_job_url_to_db, save_job_details_to_db


def init_crawler(company: str, job_type: str, location: str):
    crawlers = {
        'amazon': amazon.amazon(job_type, location),
        'meta': meta.meta(job_type, location)
    }
    if company.lower() not in crawlers:
        raise ValueError('Current company not supported')
    return crawlers[company.lower()]

def main(company: str, job_type: str, location: str):
    db = get_db(company)
    crawler = init_crawler(company, job_type, location)
    jobs = crawler.get_jobs()
    print(f'{len(jobs)} Jobs Found', flush=True)

    for job in jobs:
        print(f'Now scraping job: {job['url']}', flush=True)
        job_id = crawler.get_job_id_by_url(job['url'])
        if not job_exists(db, job_id):
            save_job_url_to_db(db, job_id, job['url'])
            details = crawler.get_job_details(job['url'])
            details["crawled_datetime"] = datetime.datetime.now().strftime("%m/%d/%Y, %H:%M:%S")
            save_job_details_to_db(db, job_id, details)

if __name__ == "__main__":
    parser = argparse.ArgumentParser(description='Job Crawler Script')
    parser.add_argument('--job_type', type=str, required=True, help='Job type for the job search')
    parser.add_argument('--location', type=str, required=True, help='Location for the job search')
    parser.add_argument('--company', type=str, required=True, help='Company for the job search')
    args = parser.parse_args()
    main(args.company, args.job_type, args.location)
