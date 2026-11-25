package model

type Exercise struct {
	ID          int     `db:"id" json:"id"`
	Name        string  `db:"name" json:"name"`
	MuscleGroup *string `db:"muscle_group" json:"muscle_group"` // Указатель, т.к. для кардио там NULL
	Type        string  `db:"type" json:"type"`
	IsCustom    bool    `db:"is_custom" json:"is_custom"`
	CreatedBy   *int    `db:"created_by" json:"created_by"` // Указатель, т.к. может быть NULL
}
