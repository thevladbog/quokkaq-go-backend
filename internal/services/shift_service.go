package services

import (
	"math"
	"quokkaq-go-backend/internal/models"
	"quokkaq-go-backend/internal/repository"
	"time"
)

type ShiftService interface {
	GetDashboardStats(unitID string) (map[string]interface{}, error)
	GetQueueTickets(unitID string) ([]models.Ticket, error)
	GetShiftCounters(unitID string) ([]ShiftCounterDTO, error)
	ExecuteEndOfDay(unitID string, userID *string) (map[string]interface{}, error)
}

type ShiftCounterDTO struct {
	models.Counter
	IsOccupied   bool           `json:"isOccupied"`
	ActiveTicket *models.Ticket `json:"activeTicket,omitempty"`
}

type shiftService struct {
	ticketRepo  repository.TicketRepository
	counterRepo repository.CounterRepository
}

func NewShiftService(ticketRepo repository.TicketRepository, counterRepo repository.CounterRepository) ShiftService {
	return &shiftService{
		ticketRepo:  ticketRepo,
		counterRepo: counterRepo,
	}
}

func (s *shiftService) GetDashboardStats(unitID string) (map[string]interface{}, error) {
	activeCountersCount, err := s.counterRepo.CountActive(unitID)
	if err != nil {
		return nil, err
	}

	queueLength, err := s.ticketRepo.CountWaiting(unitID)
	if err != nil {
		return nil, err
	}

	waitingTickets, err := s.ticketRepo.GetWaitingTickets(unitID)
	if err != nil {
		return nil, err
	}

	var averageWaitTimeMinutes float64
	if len(waitingTickets) > 0 {
		now := time.Now()
		var totalWaitMs int64
		for _, ticket := range waitingTickets {
			totalWaitMs += now.Sub(ticket.CreatedAt).Milliseconds()
		}
		averageWaitTimeMinutes = math.Round(float64(totalWaitMs) / float64(len(waitingTickets)) / 60000)
	}

	return map[string]interface{}{
		"activeCountersCount":    activeCountersCount,
		"queueLength":            queueLength,
		"averageWaitTimeMinutes": averageWaitTimeMinutes,
	}, nil
}

func (s *shiftService) GetQueueTickets(unitID string) ([]models.Ticket, error) {
	return s.ticketRepo.GetWaitingTickets(unitID)
}

func (s *shiftService) GetShiftCounters(unitID string) ([]ShiftCounterDTO, error) {
	counters, err := s.counterRepo.FindAllByUnit(unitID)
	if err != nil {
		return nil, err
	}

	dtos := make([]ShiftCounterDTO, len(counters))
	for i, counter := range counters {
		dto := ShiftCounterDTO{
			Counter: counter,
		}

		if counter.AssignedTo != nil {
			dto.IsOccupied = true
			// Get active ticket
			activeTicket, err := s.ticketRepo.GetActiveTicketByCounter(counter.ID)
			if err == nil && activeTicket != nil {
				dto.ActiveTicket = activeTicket
			}
		}

		dtos[i] = dto
	}

	return dtos, nil
}

func (s *shiftService) ExecuteEndOfDay(unitID string, userID *string) (map[string]interface{}, error) {
	// 1. Mark all active tickets (called, in_service) as served
	activeTicketsClosed, err := s.ticketRepo.UpdateStatusByUnit(unitID, []string{"called", "in_service"}, "served")
	if err != nil {
		return nil, err
	}

	// 2. Mark all waiting tickets as no_show
	waitingTicketsNoShow, err := s.ticketRepo.UpdateStatusByUnit(unitID, []string{"waiting"}, "no_show")
	if err != nil {
		return nil, err
	}

	// 3. Release all counters
	countersReleased, err := s.counterRepo.ReleaseAll(unitID)
	if err != nil {
		return nil, err
	}

	// 4. Reset ticket number sequences
	today := time.Now().Format("2006-01-02")
	err = s.ticketRepo.ResetSequences(unitID, today)
	if err != nil {
		return nil, err
	}

	// TODO: Create audit log (skipping for now as AuditLogRepo is not fully set up)

	return map[string]interface{}{
		"success":              true,
		"activeTicketsClosed":  activeTicketsClosed,
		"waitingTicketsNoShow": waitingTicketsNoShow,
		"countersReleased":     countersReleased,
	}, nil
}
