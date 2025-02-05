package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	ChatNotStarted = "not started"
	ChatProcessing = "processing"
	ChatCompleted  = "completed"
	ChatTerminated = "terminated"
)

const (
	PilotRole      = "pilot"
	SpecialistRole = "specialist"
)

const (
	RapportPhase = "rapport"
	ExplorePhase = "explore"
	EndPhase     = "end"
)

type Chat struct {
	ID                primitive.ObjectID `json:"id,omitempty" bson:"_id"`
	Messages          []Message          `json:"messages"`
	Feedback          Feedback           `json:"feedback"`
	SummarySheet      SummarySheet       `json:"summarySheet"`
	PromptVersion     int                `json:"promptVersion,omitempty"`
	ProvidePromptData ProvidePromptData  `json:"providePromptData"`
	Status            string             `json:"status"`
	StartTime         time.Time          `json:"startTime"`
	Role              string             `json:"role"`
	SessionId         string             `json:"sessionId"`
	BatchId           string             `json:"batchId"`
	UpdatedAt         time.Time          `json:"updatedAt"`
}

type BackupChat struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id"`
	Chats     []Chat             `json:"chat"`
	UpdatedAt time.Time          `json:"updatedAt"`
}

type Message struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id"`
	Role      string             `json:"role"`
	Message   string             `json:"message"`
	CreatedAt time.Time          `json:"createAt"`
	Reasoning string             `json:"reasoning"`
	Phase     string             `json:"phase"`
	Feedback  Feedback           `json:"feedback"`
}

// TODO: need to update associating to data science model
type Feedback struct {
	ID      primitive.ObjectID `json:"id,omitempty" bson:"_id"`
	Status  bool               `json:"status"`
	Message string             `json:"message"`
	Score   int                `json:"score"`
}

type ProvidePromptData struct {
	SummaryState             bool      `json:"summaryState" bson:"_id"`
	Turn                     int       `json:"turn"`
	TimeLastMessage          time.Time `json:"timeLastMessage"`
	NumberOfSelectedQuestion int       `json:"numberOfSelectedQuestion"`
}

type Session struct {
	ID           primitive.ObjectID `json:"id" bson:"_id"`
	Messages     []Message          `json:"messages"`
	Feedback     Feedback           `json:"feedback"`
	SummarySheet SummarySheet       `json:"summarySheet"`
	ChatId       string             `json:"chatId"`
	BatchId      string             `json:"batchId"`
	UpdatedAt    time.Time          `json:"updatedAt"`
}

// ServiceRecord combines basic and insight information.
type SummarySheet struct {
	Basic   BasicInfo   `json:"basic_info"`
	Insight InsightInfo `json:"insight_info"`
}

// BasicInfo represents the basic information of a service recipient.
type BasicInfo struct {
	Name        string    `json:"name"`         // ชื่อผู้รับบริการ
	Gender      string    `json:"gender"`       // เพศ
	Age         int       `json:"age"`          // อายุ
	ServiceDate time.Time `json:"service_date"` // วันที่รับบริการ
}

// InsightInfo represents the detailed insight information of the recipient.
type InsightInfo struct {
	Issue              string `json:"issue"`               // ปัญหา/เรื่องที่นำมา
	Duration           string `json:"duration"`            // ระยะเวลาที่เรื่องนี้เกิดขึ้น
	Feelings           string `json:"feelings"`            // ความรู้สึกที่มีต่อเรื่องนี้
	CopingBehaviors    string `json:"coping_behaviors"`    // พฤติกรรม/วิธีการรับมือที่มีต่อเรื่องนี้
	PhysicalSymptoms   string `json:"physical_symptoms"`   // อาการทางกายต่อเรื่องนี้
	OtherSymptoms      string `json:"other_symptoms"`      // อาการอื่นๆ
	ChronicIllness     string `json:"chronic_illness"`     // โรคประจำตัว
	PsychiatricHistory string `json:"psychiatric_history"` // ประวัติการรักษาทางจิตเวชก่อนหน้า
}
