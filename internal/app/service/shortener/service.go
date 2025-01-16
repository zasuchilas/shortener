package shortener

import (
	"context"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/zasuchilas/shortener/internal/app/logger"
	"github.com/zasuchilas/shortener/internal/app/model"
	"github.com/zasuchilas/shortener/internal/app/repository"
	"github.com/zasuchilas/shortener/internal/app/secure"
	def "github.com/zasuchilas/shortener/internal/app/service"
)

var _ def.ShortenerService = (*service)(nil)

// Group deletion settings.
const (
	DeletingChanBuffer     = 1024
	DeletingMaxRowsRequest = 512
	DeletingFlushInterval  = 10 * time.Second
)

type service struct {
	shortenerRepo repository.IStorage
	secure        *secure.Secure
	deleteCh      chan model.DeleteTask
}

// NewService _
func NewService(shortenerRepo repository.IStorage, secure *secure.Secure) *service {
	s := service{
		shortenerRepo: shortenerRepo,
		secure:        secure,
	}

	// batch deleting
	s.deleteCh = make(chan model.DeleteTask, DeletingChanBuffer)
	go s.flushDeletingTasks()

	return &s
}

// flushDeletingTasks start batch deleting urls.
func (s *service) flushDeletingTasks() {

	// the interval for sending data to the database
	ticker := time.NewTicker(DeletingFlushInterval)

	var shortURLs []string
	// TODO: use generator & buffer chan for limit shortURLs slice
	// channel for closing
	//doneCh := make(chan struct{})
	//defer close(doneCh)

	for {
		select {
		case task := <-s.deleteCh:
			shortURLs = append(shortURLs, task.ShortURLs...)
			//inputCh := deleteGenerator(doneCh, task)
		case <-ticker.C:
			// if there is nothing to send, we do not send anything
			if len(shortURLs) == 0 {
				continue
			}

			err := s.shortenerRepo.DeleteURLs(context.TODO(), shortURLs...)
			if err != nil {
				logger.Log.Info("cannot delete urls",
					zap.String("error", err.Error()), zap.String("shortURLs", strings.Join(shortURLs, ", ")))

				// we will try to delete the data next time
				continue
			}

			// clearing the deletion queue
			shortURLs = nil
		}
	}
}
