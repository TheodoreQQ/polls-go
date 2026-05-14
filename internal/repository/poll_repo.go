package repository

import (
	"database/sql"
	"errors"
	"time"

	"github.com/TheodoreQQ/polls-go/internal/models"
	"github.com/TheodoreQQ/polls-go/internal/utils"
)

// errors to handle different situations
var (
	ErrNotFound           = errors.New("record not found")
	ErrNoPermission       = errors.New("no permission to perform this action")
	ErrTxFailed           = errors.New("transaction failed")
	ErrOpFailed           = errors.New("operation failed")
	ErrAlreadyDeactivated = errors.New("poll is already deactivated")
	ErrUserAlreadyExists  = errors.New("user already exists")
)

type PollRepository struct {
	DB *sql.DB
}

func NewPollRepository(db *sql.DB) *PollRepository {
	return &PollRepository{DB: db}
}

// inserts a new poll record and its options into the db
func (r *PollRepository) Create(poll *models.ReponseForUser, userID int) error {

	tx, err := r.DB.Begin()
	if err != nil {
		return ErrTxFailed
	}

	defer tx.Rollback()

	queryPoll := `INSERT INTO polls (question, owner_id) VALUES ($1, $2) RETURNING id, created_at`
	err = tx.QueryRow(queryPoll, poll.Question, userID).Scan(&poll.ID, &poll.CreatedAt)
	if err != nil {
		return ErrOpFailed
	}

	for i := range poll.Options {
		queryOption := `INSERT INTO options (poll_id, text) VALUES ($1, $2) RETURNING id`
		err := tx.QueryRow(queryOption, poll.ID, poll.Options[i].Text).Scan(&poll.Options[i].ID)
		if err != nil {
			return ErrOpFailed
		}
		poll.Options[i].PollID = poll.ID
	}
	return tx.Commit()
}

// retrieves all polls from the database
func (r *PollRepository) GetPoll(userID int) ([]models.ReponseForUser, error) {

	queryPoll := `SELECT p.id, p.question, p.is_active, p.created_at,COALESCE(o.id, 0), COALESCE(o.text, ''),
	COALESCE(o.votes_count, 0)
								FROM polls p
								LEFT JOIN options o ON p.id = o.poll_id
								WHERE p.owner_id = $1
								ORDER BY p.id`

	rows, err := r.DB.Query(queryPoll, userID)
	if err != nil {
		return nil, ErrNotFound
	}

	defer rows.Close()

	pollsMap := make(map[int]*models.ReponseForUser)

	var pollsOrder []int

	for rows.Next() {
		var pID, oID, oVotes int
		var pQuestion, oText string
		var pActive bool
		var pCreatedAt time.Time

		err := rows.Scan(&pID, &pQuestion, &pActive, &pCreatedAt, &oID, &oText, &oVotes)
		if err != nil {
			continue
		}

		if _, exists := pollsMap[pID]; !exists {
			pollsMap[pID] = &models.ReponseForUser{
				ID:        pID,
				Question:  pQuestion,
				IsActive:  pActive,
				CreatedAt: pCreatedAt,
				Options:   []models.Option{},
			}
			pollsOrder = append(pollsOrder, pID)
		}
		if oID > 0 {
			pollsMap[pID].Options = append(pollsMap[pID].Options, models.Option{
				ID:     oID,
				PollID: pID,
				Text:   oText,
				Votes:  oVotes,
			})
		}
	}

	result := make([]models.ReponseForUser, 0)
	for _, id := range pollsOrder {
		result = append(result, *pollsMap[id])
	}

	return result, nil
}

// changes poll status from false -> true
func (r *PollRepository) Activate(pollID, userID int) (string, error) {

	roomCode := utils.GenerateCode()

	tx, err := r.DB.Begin()
	if err != nil {
		return "", ErrTxFailed
	}

	defer tx.Rollback()

	_, err = tx.Exec(`UPDATE polls SET is_active = false, code = NULL WHERE owner_id = $1`, userID)
	if err != nil {
		return "", ErrOpFailed
	}

	result, err := tx.Exec(`UPDATE polls SET is_active = true, code = $1 WHERE id = $2 AND owner_id = $3 `, roomCode, pollID, userID)

	if err != nil {
		return "", ErrOpFailed
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return "", ErrNotFound
	}

	if err := tx.Commit(); err != nil {
		return "", ErrTxFailed

	}
	return roomCode, nil
}

// retrieves a question with associated options to a student
func (r *PollRepository) GetPollByCode(code string) (*models.ResponseForStudent, error) {
	query := `SELECT p.id, p.question, o.id, o.text
						FROM polls p
						JOIN options o ON p.id = o.poll_id
						WHERE p.code = $1 AND p.is_active = true`

	rows, err := r.DB.Query(query, code)
	if err != nil {
		return nil, ErrNotFound
	}

	defer rows.Close()

	poll := &models.ResponseForStudent{}
	first := true

	for rows.Next() {
		var pID, oID int
		var pQuestion, oText string

		if err := rows.Scan(&pID, &pQuestion, &oID, &oText); err != nil {
			continue
		}

		if first {
			poll.ID = pID
			poll.Question = pQuestion
			poll.Options = []models.OptionsForStudent{}
			first = false
		}

		poll.Options = append(poll.Options, models.OptionsForStudent{
			ID:   oID,
			Text: oText,
		})
	}
	if first {
		return nil, ErrNotFound
	}

	return poll, nil
}

// increments the vote count for a specific option
func (r *PollRepository) VotePoll(optionID int) (int, error) {
	tx, err := r.DB.Begin()
	if err != nil {
		return 0, ErrTxFailed
	}

	defer tx.Rollback()

	var pollID int

	query := `
				UPDATE options 
        SET votes_count = votes_count + 1 
        WHERE id = $1 AND poll_id IN (SELECT id FROM polls WHERE is_active = true)
        RETURNING poll_id`

	err = tx.QueryRow(query, optionID).Scan(&pollID)
	if err != nil {
		return 0, ErrOpFailed
	}

	if err := tx.Commit(); err != nil {
		return 0, ErrTxFailed
	}

	return pollID, nil
}

// retrieves a specific poll with votes count
func (r *PollRepository) GetVotesByPolll(pollID, userID int) (*models.PollResultsResponse, error) {
	query := `SELECT p.id, p.question, o.id, o.text, o.votes_count FROM polls p JOIN options o ON p.id = o.poll_id WHERE p.id = $1 AND p.owner_id = $2`

	rows, err := r.DB.Query(query, pollID, userID)
	if err != nil {
		return nil, ErrNotFound
	}

	defer rows.Close()

	var response models.PollResultsResponse
	var totalVotes int

	for rows.Next() {
		var pID, oID, oVotes int
		var pQuestion, oText string

		err = rows.Scan(&pID, &pQuestion, &oID, &oText, &oVotes)
		if err != nil {
			continue
		}

		if response.ID == 0 {
			response.ID = pID
			response.Question = pQuestion
			response.Options = []models.OptionResult{}
		}

		totalVotes += oVotes

		response.Options = append(response.Options, models.OptionResult{
			ID:         oID,
			Text:       oText,
			VotesCount: oVotes,
		})
	}

	if response.ID == 0 {
		return nil, ErrNotFound
	}

	response.TotalVotes = totalVotes

	return &response, nil
}

// deletes a specific poll using its id
func (r *PollRepository) DeletePoll(pollID, userID int) error {

	query := `DELETE FROM polls WHERE id = $1 AND owner_id = $2`

	result, err := r.DB.Exec(query, pollID, userID)
	if err != nil {
		return ErrOpFailed
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// changes status from true -> false
func (r *PollRepository) DeactivatePoll(pollID, userID int) error {
	var ownerID int
	var isActive bool

	err := r.DB.QueryRow(`SELECT owner_id, is_active FROM polls WHERE id = $1`, pollID).Scan(&ownerID, &isActive)

	if err != nil {
		return ErrNotFound
	}

	if ownerID != userID {
		return ErrNoPermission
	}

	if !isActive {
		return ErrAlreadyDeactivated
	}
	result, err := r.DB.Exec("UPDATE polls SET is_active = false, code = NULL WHERE id = $1", pollID)
	if err != nil {
		return ErrOpFailed
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// allows to change question other options to a specific poll
func (r *PollRepository) UpdateQuestion(pollID, userID int, req models.UpdatePollWithOptionRequest) error {

	tx, err := r.DB.Begin()
	if err != nil {
		return ErrTxFailed
	}

	defer tx.Rollback()

	query := `UPDATE polls SET question = $1 WHERE id = $2 AND owner_id = $3`

	result, err := tx.Exec(query, req.Question, pollID, userID)

	if err != nil {
		return ErrOpFailed
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrNotFound
	}

	for _, opt := range req.Options {
		queryOpt := `UPDATE options SET text = $1 WHERE id = $2 AND poll_id = $3`
		result, err := tx.Exec(queryOpt, opt.Text, opt.ID, pollID)
		if err != nil {
			return ErrOpFailed
		}
		if rOpt, _ := result.RowsAffected(); rOpt == 0 {
			return ErrNotFound
		}
	}
	return tx.Commit()
}

// retrieves the parent poll id associated with a specific option
func (r *PollRepository) GetPollIDByOption(optionID int) (int, error) {
	var pollID int
	query := `SELECT poll_id FROM options WHERE id = $1`

	err := r.DB.QueryRow(query, optionID).Scan(&pollID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, ErrNotFound
		}
		return 0, err
	}

	return pollID, nil
}

// fetches an aggregate summary of poll results for websocket broadcasting
func (r *PollRepository) GetResultsForBroadcast(pollID int) (*models.PollResultsResponse, error) {
	var resp models.PollResultsResponse
	total := 0
	err := r.DB.QueryRow(`
		SELECT id, question, owner_id 
		FROM polls 
		WHERE id = $1`,
		pollID,
	).Scan(&resp.ID, &resp.Question, &resp.OwnerID)

	if err != nil {
		return nil, err
	}

	rows, err := r.DB.Query(`
		SELECT id, text, votes_count
		FROM options 	
		WHERE poll_id = $1 
		ORDER BY id ASC`,
		pollID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var opt models.OptionResult
		if err := rows.Scan(&opt.ID, &opt.Text, &opt.VotesCount); err != nil {
			return nil, err
		}
		total += opt.VotesCount
		resp.Options = append(resp.Options, opt)
	}

	resp.TotalVotes = total
	return &resp, nil
}
