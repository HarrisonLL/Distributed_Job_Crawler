from pymongo import MongoClient
import os


def get_db(company):
    client = MongoClient(os.getenv("MONGOURL", "mongodb://admin:adminpass@localhost:27017"))
    db = client[f'{company}_jobcrawler']
    return db

def job_exists(db, job_id):
    jobs_collection = db['jobs']
    return jobs_collection.find_one({'id': job_id}) is not None

def save_job_url_to_db(db, job_id, url):
    jobs_collection = db['jobs']
    init_dict = {'id': job_id, 'url': url}
    jobs_collection.insert_one(init_dict)

def save_job_details_to_db(db, job_id, job_details, update = True):
    '''
    params: 
        update: True: update existing job data; False: insert job data
    '''
    jobs_collection = db['jobs']
    if update:
        jobs_collection.update_one({'id': job_id}, {'$set': job_details})
    else:
        jobs_collection.insert_one(job_details)

def get_job_type(db, job_id):
    jobs_collection = db['jobs']
    job = jobs_collection.find_one({'id': job_id})
    return job['job_types'] if job and 'job_types' in job else []

def save_job_type_to_db(db, job_id, job_type, updated_time):
    jobs_collection = db['jobs']
    job_details = jobs_collection.find_one({'id': job_id})
    job_details['job_types'] = job_type
    job_details['crawled_datetime'] = updated_time
    jobs_collection.update_one({'id': job_id}, {'$set': job_details})
