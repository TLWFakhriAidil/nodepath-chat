package models

import (
	"database/sql"
)

// WasapBot represents the wasapBot_nodepath table structure
type WasapBot struct {
	IDProspect       int            `json:"id_prospect"`
	FlowReference    sql.NullString `json:"flow_reference"`
	ExecutionID      sql.NullString `json:"execution_id"`
	ExecutionStatus  sql.NullString `json:"execution_status"`
	FlowID           sql.NullString `json:"flow_id"`
	CurrentNodeID    sql.NullString `json:"current_node_id"`
	LastNodeID       sql.NullString `json:"last_node_id"`
	WaitingForReply  int            `json:"waiting_for_reply"`
	IDDevice         sql.NullString `json:"id_device"`
	ProspectNum      sql.NullString `json:"prospect_num"`
	Niche            sql.NullString `json:"niche"`
	PeringkatSekolah sql.NullString `json:"peringkat_sekolah"`
	Alamat           sql.NullString `json:"alamat"`
	Nama             sql.NullString `json:"nama"`
	Pakej            sql.NullString `json:"pakej"`
	NoFon            sql.NullString `json:"no_fon"`
	CaraBayaran      sql.NullString `json:"cara_bayaran"`
	TarikhGaji       sql.NullString `json:"tarikh_gaji"`
	Stage            sql.NullString `json:"stage"`
	TempStage        sql.NullString `json:"temp_stage"`
	ConvStart        sql.NullString `json:"conv_start"`
	ConvLast         sql.NullString `json:"conv_last"`
	DateStart        sql.NullString `json:"date_start"`
	DateLast         sql.NullString `json:"date_last"`
	Status           sql.NullString `json:"status"`
	StaffCls         sql.NullString `json:"staff_cls"`
	Umur             sql.NullString `json:"umur"`
	Kerja            sql.NullString `json:"kerja"`
	Sijil            sql.NullString `json:"sijil"`
	UserInput        sql.NullString `json:"user_input"`
	Alasan           sql.NullString `json:"alasan"`
	Nota             sql.NullString `json:"nota"`
}
