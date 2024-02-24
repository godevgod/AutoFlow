//if you interest please donate 13w5Hyjc7CzGnCpKte4M4narsTZoEA8QX7
//then make a contact me godevgod to improve this code

package main

import (
	"fmt"
	"log"
	"sync/atomic"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Job holds details about a job.
//
//	type Job struct {
//		ID          int64
//		Type        string
//		Description string
//		Status      atomic.Value // Manages the job's status atomically.
//	}
type Job struct {
	ID          int64
	Type        string
	Description string
	Status      atomic.Value           // Manages the job's status atomically.
	Parameters  map[string]interface{} // Stores job-specific parameters.
	Trigger     atomic.Value           // Indicates specific trigger conditions or commands.
}

type JobWatcher struct {
	nextJobID int64
	jobs      atomic.Value // Stores map[int64]*AtomicJob
}

// NewJob initializes a new job with specified details.
func NewJob(id int64, jobType, description string) *Job {
	job := &Job{ID: id, Type: jobType, Description: description}
	job.Status.Store("new") // Initializes status to "new".
	return job
}

// AtomicJob provides an atomic reference to a Job.
type AtomicJob struct {
	value atomic.Value
}

func (aj *AtomicJob) Store(job *Job) {
	aj.value.Store(job)
}

func (aj *AtomicJob) Load() *Job {
	if v := aj.value.Load(); v != nil {
		return v.(*Job)
	}
	return nil
}

func NewJobWatcher() *JobWatcher {
	jw := &JobWatcher{}
	jw.jobs.Store(make(map[int64]*AtomicJob))
	return jw
}

func (jw *JobWatcher) AddJob(jobType, description string) {
	jobID := atomic.AddInt64(&jw.nextJobID, 1)
	job := &Job{ID: jobID, Type: jobType, Description: description}
	job.Status.Store("new")
	atomicJob := &AtomicJob{}
	atomicJob.Store(job)

	// Safely update the map
	jobsMap := jw.jobs.Load().(map[int64]*AtomicJob)
	newMap := make(map[int64]*AtomicJob)
	for k, v := range jobsMap {
		newMap[k] = v
	}
	newMap[jobID] = atomicJob
	jw.jobs.Store(newMap)

	fmt.Printf("Added job: ID=%d, Type=%s\n", jobID, jobType)
}

func SetupRoutes(app *fiber.App, watcher *JobWatcher) {
	app.Post("/newjob", func(c *fiber.Ctx) error {
		jobType := c.FormValue("type")
		description := c.FormValue("description")
		watcher.AddJob(jobType, description)
		return c.JSON(fiber.Map{"message": "Job added successfully"})
	})
}

func AutonomousObserver(watcher *JobWatcher, condition string, action func(*Job), triggerCondition string) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		jobsMap := watcher.jobs.Load().(map[int64]*AtomicJob)
		for _, aj := range jobsMap {
			job := aj.Load()
			if job != nil {
				// ตรวจสอบก่อนว่า Trigger ไม่เป็น nil
				triggerValue := job.Trigger.Load()
				var trigger string
				if triggerValue != nil {
					trigger, _ = triggerValue.(string) // ใช้ comma-ok idiom เพื่อป้องกัน panic
				}

				// ตรวจสอบต่อว่า triggerCondition ตรงกับที่ระบุหรือไม่ (ถ้ามีการระบุ)
				// และสถานะของงานตรงกับ condition
				if (triggerCondition == "" || trigger == triggerCondition) && job.Status.Load() == condition {
					action(job)
				}
			}
		}
	}
}

func main() {
	app := fiber.New()
	watcher := NewJobWatcher()

	SetupRoutes(app, watcher)
	ProcessNewJobs(watcher)
	PerformQualityAssurance(watcher)

	FinalApprovalProcess(watcher)

	log.Fatal(app.Listen(":8080"))

}

func ProcessNewJobs(watcher *JobWatcher) {
	action := func(job *Job) {
		fmt.Printf("Processing job: ID=%d\n", job.ID)
		job.Status.Store("review")
	}
	// ใช้ค่าว่างสำหรับ triggerCondition เนื่องจากไม่มีเงื่อนไขเฉพาะ
	go AutonomousObserver(watcher, "new", action, "")
}

// PerformQualityAssurance conducts QA on jobs marked for review.
func PerformQualityAssurance(watcher *JobWatcher) {
	qaAction := func(job *Job) {
		fmt.Printf("QA on job: ID=%d\n", job.ID)
		// Replace with actual QA logic
		job.Status.Store("approved")
	}
	go AutonomousObserver(watcher, "review", qaAction, "")
}

func FinalApprovalProcess(watcher *JobWatcher) {
	finalApprovalAction := func(job *Job) {
		// แสดงข้อมูลงานที่กำลังรอการอนุมัติสุดท้าย
		fmt.Printf("Final approval process started for job: ID=%d, Type=%s\n", job.ID, job.Type)

		// ทำการประมวลผลการอนุมัติสุดท้าย ซึ่งอาจรวมถึงการตรวจสอบเงื่อนไขพิเศษ, การตรวจสอบความถูกต้องของข้อมูล ฯลฯ
		// สำหรับตัวอย่าง, เราจะใช้เวลา 2 วินาทีในการจำลองการตรวจสอบ
		time.Sleep(2 * time.Second)

		// อัปเดตสถานะของงานเป็น "completed" หลังจากการอนุมัติสุดท้าย
		job.Status.Store("completed")
		fmt.Printf("Final approval completed for job: ID=%d, new Status=%s\n", job.ID, "completed")
	}

	// เริ่มต้น AutonomousObserver สำหรับการติดตามและดำเนินการกับงานที่มีสถานะ "approved"
	go AutonomousObserver(watcher, "approved", finalApprovalAction, "")
}

func FireAlertProcess(watcher *JobWatcher) {
	fireAlertAction := func(job *Job) {
		// ตรวจสอบเงื่อนไขอุณหภูมิจาก Parameters
		if temperature, ok := job.Parameters["temperature"].(int); ok && temperature > 100 {
			fmt.Printf("Fire alert for job: ID=%d, Temperature: %d\n", job.ID, temperature)
			job.Status.Store("fire_alert_handled")
		}
	}

	// ใช้ Trigger "fire_alert" เพื่อเรียกใช้งานนี้
	go AutonomousObserver(watcher, "new", fireAlertAction, "fire_alert")
}

/*

Refined Function Template
Here is a polished version of the function template that adheres to Go's idiomatic practices:


func WatchAndProcessJobs(watcher *JobWatcher, triggerStatus string, process func(job *Job)) {
    processJob := func(job *Job) {
        // Explain what processing will occur here, specific to the triggerStatus.
        // For example, "Processing job for QA checks" for a triggerStatus of "review".
        fmt.Printf("Triggered processing on job: ID=%d, Type=%s, Status=%s\n", job.ID, job.Type, triggerStatus)

        // Simulate some processing time.
        time.Sleep(2 * time.Second)

        // Update job's status after processing.
        // For instance, setting status to "completed" after a "new" job is processed.
        job.Status.Store("completed")
        fmt.Printf("Processed job: ID=%d, new Status=%s\n", job.ID, "completed")
    }

    // Start the AutonomousObserver goroutine to monitor and act upon jobs with the specified triggerStatus.
    go AutonomousObserver(watcher, triggerStatus, processJob)
}
*/

// การทำงานอัตโนมัติของฟังก์ชันต่างๆ:
// - ProcessNewJobs: ดำเนินการกับงานที่มีสถานะ "new" โดยอัตโนมัติ ไม่ต้องการ triggerCondition เฉพาะเจาะจง
// - PerformQualityAssurance: ดำเนินการตรวจสอบคุณภาพ (QA) กับงานที่มีสถานะ "review" โดยอัตโนมัติ
// - FinalApprovalProcess: ดำเนินการอนุมัติสุดท้ายกับงานที่มีสถานะ "approved" โดยอัตโนมัติ
// - FireAlertProcess: ตอบสนองต่อสถานการณ์ไฟไหม้ (เช่น อุณหภูมิเกิน 100 องศา) โดยต้องมีการระบุ triggerCondition เป็น "fire_alert"

// การรับเงื่อนไขพิเศษ:
// - เงื่อนไขพิเศษสามารถรับได้โดยการใช้ Trigger ซึ่งเป็น atomic.Value ใน Job ที่เก็บค่าเงื่อนไขเฉพาะหรือคำสั่งที่ระบุโดย JobWatcher หรือระบบภายนอก
// - โดยการตั้งค่า Trigger ให้กับงานเฉพาะ, คุณสามารถกำหนดให้ AutonomousObserver ทำงานกับงานนั้นๆ เมื่อ Trigger ตรงกับ triggerCondition ที่ระบุในการเรียกใช้ AutonomousObserver
// - สำหรับสถานการณ์เช่นไฟไหม้, คุณอาจใช้ Parameters เพื่อเก็บข้อมูลอุณหภูมิ และตั้งค่า Trigger เมื่ออุณหภูมิเกินค่าที่กำหนด

// ประโยชน์:
// - การตอบสนองอัตโนมัติ: ระบบสามารถตอบสนองต่อเหตุการณ์ต่างๆ โดยอัตโนมัติ ไม่ว่าจะเป็นการประมวลผลงานใหม่, การตรวจสอบ QA, หรือการดำเนินการฉุกเฉิน
// - ความยืดหยุ่น: ผ่านการใช้ Trigger และ Parameters, ระบบสามารถรับมือกับสถานการณ์ที่หลากหลายและเฉพาะเจาะจงได้

// AutonomousObserver: ฟังก์ชันที่เป็นหัวใจของระบบการตอบสนองอัตโนมัติ

// ความสามารถหลัก:
// - ตรวจสอบงานในระบบอย่างต่อเนื่องตามเงื่อนไขที่กำหนด (เช่น สถานะของงาน).
// - ตอบสนองต่อเหตุการณ์หรือเงื่อนไขพิเศษโดยการใช้ Trigger ที่กำหนดในงาน.

// การทำงาน:
// 1. ใช้ ticker ที่ตั้งเวลาเป็น 1 วินาทีเพื่อสร้างลูปการทำงานอย่างต่อเนื่อง.
// 2. โหลดรายการงานทั้งหมดจาก JobWatcher และตรวจสอบแต่ละงานว่าตรงกับเงื่อนไขที่ระบุหรือไม่.
// 3. สำหรับงานที่มี Trigger, ตรวจสอบว่า Trigger ตรงกับ triggerCondition ที่ระบุหรือไม่.
// 4. หากตรงกับเงื่อนไข, ทำการดำเนินการกับงานนั้นๆ ตามที่ฟังก์ชัน action กำหนด.

// การใช้งาน:
// - สามารถปรับใช้ในสถานการณ์ต่างๆ เช่น การประมวลผลงานใหม่, การตรวจสอบคุณภาพ, หรือการตอบสนองต่อสถานการณ์ฉุกเฉิน (เช่น ไฟไหม้).

// ประโยชน์:
// - การตอบสนองอัตโนมัติและทันทีต่อเงื่อนไขหรือเหตุการณ์ที่เปลี่ยนแปลง.
// - ความยืดหยุ่นและความสามารถในการปรับตัวสูงเนื่องจากสามารถกำหนดเงื่อนไขหรือ Trigger ได้หลากหลาย.
// - ลดความจำเป็นในการมีการแทรกแซงจากมนุษย์, เพิ่มประสิทธิภาพในการจัดการงาน.

// การใช้งานในรูปแบบต่างๆ:
// - `ProcessNewJobs`, `PerformQualityAssurance`, `FinalApprovalProcess`, และ `FireAlertProcess` ทำงานอย่างอัตโนมัติโดยมีการตรวจสอบและตอบสนองตามเงื่อนไขที่ระบุ.
// - การตอบสนองต่อสถานการณ์พิเศษ เช่น ไฟไหม้, ผ่านการตั้งค่า Trigger และ Parameters ให้เหมาะสมกับสถานการณ์นั้นๆ.

// โดยสรุป, `AutonomousObserver` เป็นฟังก์ชันที่ให้ความสามารถในการตอบสนองและจัดการงานในระบบอย่างอัตโนมัติและมีประสิทธิภาพ, ช่วยให้ระบบสามารถรับมือกับสถานการณ์และเงื่อนไขที่หลากหลายได้ดีขึ้น.

// Automated Functionality of Various Functions:
// - ProcessNewJobs: Automatically processes jobs with the "new" status without requiring specific trigger conditions.
// - PerformQualityAssurance: Automatically conducts quality assurance (QA) for jobs with the "review" status.
// - FinalApprovalProcess: Automatically carries out final approval for jobs with the "approved" status.
// - FireAlertProcess: Responds to fire situations (e.g., temperature exceeding 100 degrees) by requiring the specification of "fire_alert" as a trigger condition.

// Handling Special Conditions:
// - Special conditions can be accommodated using a Trigger, which is an atomic.Value in Job that stores specific conditions or commands specified by JobWatcher or an external system.
// - By setting a Trigger for a particular job, you can ensure the AutonomousObserver operates on that job when the Trigger matches the specified triggerCondition.
// - For scenarios like fire, you might use Parameters to store temperature data and set a Trigger when the temperature exceeds a predefined value.

// Benefits:
// - Automated Response: The system can automatically respond to various events, whether processing new jobs, conducting QA, or emergency operations.
// - Flexibility: Through the use of Triggers and Parameters, the system can handle a wide variety of specific situations.

// AutonomousObserver: The Heart of the Automated Response System

// Core Capabilities:
// - Continuously monitors jobs in the system based on specified conditions (e.g., job status).
// - Responds to special events or conditions using Triggers defined in the job.

// How It Works:
// 1. Uses a ticker set to 1-second intervals to create a continuous loop of operation.
// 2. Loads all jobs from JobWatcher and checks each job against the specified condition.
// 3. For jobs with a Trigger, checks if the Trigger matches the specified triggerCondition.
// 4. If the condition is met, performs the action defined by the function action on that job.

// Usage:
// - Can be deployed in various scenarios, such as processing new jobs, quality assurance, or responding to emergencies (like fire).

// Benefits:
// - Immediate and automated response to changing conditions or events.
// - High flexibility and adaptability due to the ability to specify diverse conditions or Triggers.
// - Reduces the need for human intervention, enhancing efficiency and speed in job management.

// Application in Different Forms:
// - `ProcessNewJobs`, `PerformQualityAssurance`, `FinalApprovalProcess`, and `FireAlertProcess` operate automatically by monitoring and responding to specified conditions.
// - Responds to special situations, like fire, through appropriate setting of Triggers and Parameters.

// In summary, `AutonomousObserver` is a function that enables efficient and automated operation and management of jobs within the system, allowing the system to better handle diverse situations and conditions without human intervention in every step.
