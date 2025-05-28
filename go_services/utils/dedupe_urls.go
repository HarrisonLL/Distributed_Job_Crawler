package utils

import (
	"sort"
	"time"
)

func DedupAndSortLinkedInJobs(jobs []JobDetail) []JobDetail {
	uniqueJobs := make(map[string]JobDetail)

	for _, job := range jobs {
		uniqueJobs[job.URL] = job
	}

	var deduped []JobDetail
	for _, job := range uniqueJobs {
		deduped = append(deduped, job)
	}

	sort.Slice(deduped, func(i, j int) bool {
		dateI, errI := time.Parse("2006-01-02", deduped[i].Qualifications)
		dateJ, errJ := time.Parse("2006-01-02", deduped[j].Qualifications)

		switch {
		case errI == nil && errJ == nil:
			if dateI.Equal(dateJ) {
				// Same posting date, sort by company
				return deduped[i].Company < deduped[j].Company
			}
			return dateI.After(dateJ)
		default:
			// fallback to string comparison
			if deduped[i].Qualifications == deduped[j].Qualifications {
				return deduped[i].Company < deduped[j].Company
			}
			return deduped[i].Qualifications > deduped[j].Qualifications
		}
	})

	return deduped
}
