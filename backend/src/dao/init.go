package dao

import (
	"fmt"
	"hr-dynamic-form-engine/src/cfg"
	"hr-dynamic-form-engine/src/log"
	"hr-dynamic-form-engine/src/model"
	"time"

	_ "github.com/lib/pq"
	"xorm.io/xorm"
)

var defaultEngine *xorm.Engine

func DBDefault() *xorm.Engine {
	if defaultEngine == nil {
		panic(fmt.Errorf("数据库连接尚未初始化"))
	}
	return defaultEngine
}

func Init() {
	if cfg.Cfg.PgsqlDSN == "" {
		panic(fmt.Errorf("PostgreSQL 配置不能为空 (cfg.Cfg.PgsqlDSN is empty)"))
	}

	log.Zap.Infof("正在连接 PostgreSQL: %s ...", maskDSN(cfg.Cfg.PgsqlDSN))
	engine, err := xorm.NewEngine("postgres", cfg.Cfg.PgsqlDSN)
	if err != nil {
		panic(fmt.Errorf("创建 PostgreSQL 引擎失败: %w", err))
	}

	maxConn := cfg.Cfg.Pgsql.MaxConn
	if maxConn <= 0 {
		maxConn = 50
	}
	idleConn := cfg.Cfg.Pgsql.IdleConn
	if idleConn <= 0 {
		idleConn = 10
	}

	engine.SetMaxOpenConns(maxConn)
	engine.SetMaxIdleConns(idleConn)
	engine.SetConnMaxLifetime(time.Hour)
	if cfg.Cfg.Env == "dev" {
		engine.ShowSQL(true)
	}

	if err = engine.Ping(); err != nil {
		panic(fmt.Errorf("PostgreSQL 数据库连接失败: %w", err))
	}

	defaultEngine = engine

	initTables(engine)
	seedInitialData(engine)

	log.Zap.Infoln("dao start success")
}

func maskDSN(dsn string) string {
	if len(dsn) > 20 {
		return dsn[:10] + "****" + dsn[len(dsn)-5:]
	}
	return dsn
}

func initTables(engine *xorm.Engine) {
	err := engine.Sync2(
		new(model.FormSchema),
		new(model.FormSchemaVersion),
		new(model.FormSubmission),
		new(model.FileMetadata),
		new(model.AuditLog),
	)
	if err != nil {
		panic(fmt.Errorf("PostgreSQL 同步数据表结构失败: %w", err))
	}

	ginIndexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_submissions_data_gin ON form_submission USING GIN (data);`,
		`CREATE INDEX IF NOT EXISTS idx_submissions_form_id ON form_submission (form_id);`,
		`CREATE INDEX IF NOT EXISTS idx_submissions_created_at ON form_submission (created_at DESC);`,
		`CREATE INDEX IF NOT EXISTS idx_audit_logs_form_id ON audit_log (form_id);`,
		`CREATE INDEX IF NOT EXISTS idx_audit_logs_target ON audit_log (target_type, target_id);`,
	}

	for _, idxSql := range ginIndexes {
		if _, err := engine.Exec(idxSql); err != nil {
			log.Zap.Warnf("创建索引提示: %v", err)
		}
	}
}

func seedInitialData(engine *xorm.Engine) {
	formId := "tech_onboarding"
	has, err := engine.Table("form_schema").Where("form_id = ?", formId).Exist()
	if err != nil {
		panic(err)
	}
	if has {
		return
	}

	minLen1 := 2
	maxLen50 := 50
	minSalary := 5000.0
	maxSalary := 200000.0
	minExp := 0.0
	maxExp := 50.0

	initialSchema := &model.FormSchema{
		Id:               "schema_tech_onboarding_v1",
		FormId:           formId,
		Title:            "技术岗位入职信息表",
		Description:      "收集技术岗新入职员工基础档案、GitHub主页、工作年限、工作履历及身份证材料",
		Version:          1,
		PublishedVersion: 1,
		Status:           model.SchemaStatusPublished,
		Migrations: map[string]string{
			"employeeName": "fullName",
		},
		CreatedBy: "HRBP_Admin",
		Fields: []model.FieldDefinition{
			{
				Key:         "fullName",
				Label:       "员工姓名",
				Type:        "input",
				Placeholder: "请输入真实姓名",
				Validation: &model.ValidationRules{
					Required:  true,
					MinLength: &minLen1,
					MaxLength: &maxLen50,
				},
				GridCol: 12,
			},
			{
				Key:        "nickname",
				Label:      "花名/昵称",
				Type:       "input",
				Deprecated: true,
				GridCol:    12,
			},
			{
				Key:         "email",
				Label:       "工作邮箱",
				Type:        "input",
				Placeholder: "developer@company.com",
				Validation: &model.ValidationRules{
					Required:      true,
					Email:         true,
					CustomMessage: "字段 email 格式不正确",
				},
				GridCol: 12,
			},
			{
				Key:         "phone",
				Label:       "联系电话",
				Type:        "input",
				Placeholder: "11位手机号码",
				Validation: &model.ValidationRules{
					Required:      true,
					Regex:         `^1[3-9]\d{9}$`,
					CustomMessage: "手机号格式必须为11位有效号码",
				},
				GridCol: 12,
			},
			{
				Key:          "department",
				Label:        "所属部门",
				Type:         "select",
				Placeholder:  "请选择入职部门",
				DefaultValue: "tech_dept",
				Validation: &model.ValidationRules{
					Required: true,
				},
				Options: []model.OptionItem{
					{Label: "基础架构部", Value: "infra"},
					{Label: "后端研发部", Value: "tech_dept"},
					{Label: "前端中台组", Value: "fe_team"},
					{Label: "AI 智能实验室", Value: "ai_lab"},
				},
				GridCol: 12,
			},
			{
				Key:         "githubUrl",
				Label:       "GitHub 个人主页",
				Type:        "input",
				Placeholder: "https://github.com/your-username",
				Validation: &model.ValidationRules{
					Required: true,
					Regex:    `^https?://github\.com/.*`,
				},
				GridCol: 12,
			},
			{
				Key:          "isFreshGraduate",
				Label:        "是否应届生",
				Type:         "radio",
				DefaultValue: "no",
				Validation: &model.ValidationRules{
					Required: true,
				},
				Options: []model.OptionItem{
					{Label: "是（应届毕业）", Value: "yes"},
					{Label: "否（社招有经验）", Value: "no"},
				},
				GridCol: 12,
				Rules: []model.FieldRule{
					{
						TargetKey:  "previousCompany",
						Action:     "hide",
						Expression: "isFreshGraduate == 'yes'",
					},
					{
						TargetKey:  "previousCompany",
						Action:     "unsetRequired",
						Expression: "isFreshGraduate == 'yes'",
					},
				},
			},
			{
				Key:         "previousCompany",
				Label:       "前任公司名称",
				Type:        "input",
				Placeholder: "社招必填：上一任就职单位",
				Validation: &model.ValidationRules{
					Required: true,
				},
				GridCol: 12,
			},
			{
				Key:          "experience",
				Label:        "工作年限 (年)",
				Type:         "number",
				DefaultValue: 3,
				Validation: &model.ValidationRules{
					Required: true,
					Min:      &minExp,
					Max:      &maxExp,
				},
				GridCol: 12,
			},
			{
				Key:          "expectedSalary",
				Label:        "期望薪资 (元/月)",
				Type:         "number",
				DefaultValue: 28000,
				Validation: &model.ValidationRules{
					Required: true,
					Min:      &minSalary,
					Max:      &maxSalary,
				},
				GridCol: 12,
				Rules: []model.FieldRule{
					{
						TargetKey:  "salary_approval_alert",
						Action:     "showAlert",
						Expression: "expectedSalary > 30000",
						Parameters: map[string]any{
							"text": "【审批警示】期望薪资超过 30,000 元/月，提交后系统将自动流转至 HRBP 总监特别审批！",
						},
					},
				},
			},
			{
				Key:   "salary_approval_alert",
				Label: "薪酬审批提示",
				Type:  "alert",
				AlertConfig: &model.AlertConfig{
					Title: "需要 HRBP 特别审批",
					Type:  "warning",
				},
				GridCol: 24,
			},
			{
				Key:   "workHistory",
				Label: "工作履历",
				Type:  "repeater",
				ItemFields: []model.FieldDefinition{
					{
						Key:         "company",
						Label:       "就职公司",
						Type:        "input",
						Placeholder: "公司全称",
						Validation: &model.ValidationRules{
							Required: true,
						},
					},
					{
						Key:         "role",
						Label:       "担任职位",
						Type:        "input",
						Placeholder: "如：高级 Golang 研发工程师",
						Validation: &model.ValidationRules{
							Required: true,
						},
					},
					{
						Key:   "startDate",
						Label: "入职时间",
						Type:  "date",
						Validation: &model.ValidationRules{
							Required: true,
						},
					},
					{
						Key:   "endDate",
						Label: "离职时间",
						Type:  "date",
					},
				},
				GridCol: 24,
			},
			{
				Key:   "emergencyContact",
				Label: "紧急联系人信息",
				Type:  "object",
				Children: []model.FieldDefinition{
					{
						Key:         "name",
						Label:       "联系人姓名",
						Type:        "input",
						Placeholder: "姓名",
						Validation: &model.ValidationRules{
							Required: true,
						},
						GridCol: 12,
					},
					{
						Key:          "relation",
						Label:        "与本人关系",
						Type:         "select",
						DefaultValue: "parents",
						Options: []model.OptionItem{
							{Label: "父母", Value: "parents"},
							{Label: "配偶", Value: "spouse"},
							{Label: "朋友", Value: "friend"},
						},
						GridCol: 12,
					},
					{
						Key:         "contactPhone",
						Label:       "紧急联系电话",
						Type:        "input",
						Placeholder: "11位手机号码",
						Validation: &model.ValidationRules{
							Required: true,
							Regex:    `^1[3-9]\d{9}$`,
						},
						GridCol: 24,
					},
				},
				GridCol: 24,
			},
			{
				Key:   "idCardCopy",
				Label: "身份证正反面影印件",
				Type:  "upload",
				Validation: &model.ValidationRules{
					Required: true,
				},
				GridCol: 12,
			},
			{
				Key:   "resumeAttachment",
				Label: "最新个人简历 (PDF/Word)",
				Type:  "upload",
				GridCol: 12,
			},
		},
	}

	_, err = engine.InsertOne(initialSchema)
	if err != nil {
		panic(err)
	}

	snap := &model.FormSchemaVersion{
		FormId:      formId,
		Version:     1,
		Title:       initialSchema.Title,
		Description: initialSchema.Description,
		Fields:      initialSchema.Fields,
		PublishedBy: "HRBP_Admin",
	}
	_, _ = engine.InsertOne(snap)

	sub1 := &model.FormSubmission{
		Id:            "sub_demo_001",
		FormId:        formId,
		SchemaVersion: 1,
		SubmitterId:   "user_101",
		SubmitterName: "张小强",
		Status:        "submitted",
		Data: map[string]interface{}{
			"fullName":        "张小强",
			"email":           "zhangxq@example.com",
			"phone":           "13800138000",
			"department":      "tech_dept",
			"githubUrl":       "https://github.com/zhangxq-dev",
			"isFreshGraduate": "no",
			"previousCompany": "字节跳动",
			"experience":      5,
			"expectedSalary":  35000,
			"workHistory": []interface{}{
				map[string]interface{}{
					"company":   "字节跳动",
					"role":      "高级Go后端开发",
					"startDate": "2021-03-01",
					"endDate":   "2024-05-30",
				},
			},
			"emergencyContact": map[string]interface{}{
				"name":         "张大伯",
				"relation":     "parents",
				"contactPhone": "13912345678",
			},
			"idCardCopy": map[string]interface{}{
				"fileId":      "file_demo_idcard_01",
				"name":        "id_card_front.jpg",
				"size":        1024 * 350,
				"downloadUrl": "/api/v1/files/download/file_demo_idcard_01",
			},
		},
	}

	sub2 := &model.FormSubmission{
		Id:            "sub_demo_002",
		FormId:        formId,
		SchemaVersion: 1,
		SubmitterId:   "user_102",
		SubmitterName: "李思思",
		Status:        "submitted",
		Data: map[string]interface{}{
			"fullName":        "李思思",
			"email":           "lisisi@example.com",
			"phone":           "13900139000",
			"department":      "fe_team",
			"githubUrl":       "https://github.com/lisisi-fe",
			"isFreshGraduate": "yes",
			"experience":      1,
			"expectedSalary":  22000,
			"workHistory":     []interface{}{},
			"emergencyContact": map[string]interface{}{
				"name":         "李母",
				"relation":     "parents",
				"contactPhone": "13700137000",
			},
			"idCardCopy": map[string]interface{}{
				"fileId":      "file_demo_idcard_02",
				"name":        "id_card_lisi.png",
				"size":        1024 * 280,
				"downloadUrl": "/api/v1/files/download/file_demo_idcard_02",
			},
		},
	}
	_, _ = engine.InsertOne(sub1)
	_, _ = engine.InsertOne(sub2)

	auditLog := &model.AuditLog{
		Action:       "UPDATE_SUBMISSION",
		TargetType:   "SUBMISSION",
		TargetId:     sub1.Id,
		FormId:       formId,
		OperatorId:   "hr_user_007",
		OperatorName: "HRBP 评审专员",
		ClientIp:     "192.168.1.108",
		Diff: map[string]model.DiffItem{
			"expectedSalary": {Old: 30000, New: 35000},
			"department":     {Old: "fe_team", New: "tech_dept"},
		},
	}
	_, _ = engine.InsertOne(auditLog)
}
