package hashcash

import (
	"bytes"
	"context"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Our version of the header consists of 7 fields
//
// HashCashHeader: version:difficult:expiredAt:subject:alg:nonce:counter
//
///////////////////////////////////////////////////////////////////////////////////////////
// +------------+-------------+-------------+-----------+-------+--------+------------+ //
// |  version  |  difficult  |  expiredAt  |  subject  |  alg  |  nonce  |  counter  | //
//+-----------+-------------+-------------+-----------+-------+---------+-----------+ //
///////////////////////////////////////////////////////////////////////////////////////
// version 		-> version of the header, represented by an int
// difficult 	-> the complexity of the useful work, expressed in the number of first zeros of the hash
// expiredAt	-> the lifetime of the task, after which it is considered invalid
// subject 		-> defines the name of the resource or other identifying information, such as a user ID or its ip address
// alg			-> determines what type of hash to use
// nonce		-> is a randomly generated set of bytes
// counter 		-> Current counter value
// Example 1:20:1665396610:localhost:sha-512:hVscDCMZcS1WYg==:BQAAAAAAAAA=
const (
	AlgSHA256 = "sha-256"
	AlgSHA512 = "sha-512"
)

var (
	defaultVersion      = 1
	defaultRandBytesNum = 10
	defaultHashAlg      = AlgSHA256
)

var (
	ErrHeaderInvalid = errors.New("header invalid")
	ErrMaxIterations = errors.New("max iterations")
)

type Header struct {
	Version   int
	Difficult int
	Subject   string
	ExpiredAt int64
	Alg       string
	Nonce     string
	Counter   uint64
}

func (h Header) IsValid() bool {
	return Verify(h.Hash(), h.Difficult)
}

func (h Header) Bytes() []byte {
	buf := &bytes.Buffer{}
	_, _ = fmt.Fprintf(buf, "%d:%d:%d:%s:%s:%s:%s", h.Version, h.Difficult, h.ExpiredAt, h.Subject,
		h.Alg, h.Nonce, uint64ToBase64(h.Counter))

	return buf.Bytes()
}

func (h Header) String() string {
	return fmt.Sprintf("%d:%d:%d:%s:%s:%s:%s", h.Version, h.Difficult, h.ExpiredAt, h.Subject,
		h.Alg, h.Nonce, uint64ToBase64(h.Counter))
}

func (h Header) Hash() string {
	f := hashFunc(h.Alg)
	f.Write([]byte(h.String()))

	return hex.EncodeToString(f.Sum(nil))
}

// Default create header from default parameters
func Default(subject string, difficult int, expiredAt int64) (Header, error) {
	rnd, err := randomString(defaultRandBytesNum)
	if err != nil {
		return Header{}, fmt.Errorf("randomString: %w", err)
	}

	return Header{
		Version:   defaultVersion,
		Difficult: difficult,
		Subject:   subject,
		ExpiredAt: expiredAt,
		Alg:       defaultHashAlg,
		Nonce:     rnd,
		Counter:   0,
	}, nil
}

// Compute the useful work according to the header
func Compute(ctx context.Context, header Header, maxIterations int) (Header, error) {
	counter := int(header.Counter)

	// try to find a suitable hash value
	for counter <= maxIterations || maxIterations <= 0 {
		select {
		case <-ctx.Done():
			return Header{}, ctx.Err()
		default:
		}

		if header.IsValid() {
			return header, nil
		}

		counter++
		header.Counter++
	}

	return Header{}, ErrMaxIterations
}

type PoolOption func(*PoolConfig)

type PoolConfig struct {
	WorkerNum         int
	Duration          *time.Duration
	PoolMaxIterations *int
}

type PoolInfo struct {
	Time      time.Duration
	WorkerNum int
	Header    Header
}

func WithPoolDuration(d time.Duration) PoolOption {
	return func(config *PoolConfig) {
		config.Duration = &d
	}
}

func WithPoolMaxIterations(i int) PoolOption {
	return func(config *PoolConfig) {
		config.PoolMaxIterations = &i
	}
}

func WithWorkerNum(n int) PoolOption {
	return func(config *PoolConfig) {
		config.WorkerNum = n
	}
}

const defaultPoolWorkerNum = 1

func ComputeWithPool(ctx context.Context, header Header, opt ...PoolOption) (PoolInfo, error) {
	var multiCtx context.Context
	var cancel context.CancelFunc

	config := PoolConfig{WorkerNum: defaultPoolWorkerNum}

	for _, option := range opt {
		option(&config)
	}

	if config.Duration != nil {
		multiCtx, cancel = context.WithTimeout(ctx, *config.Duration)
	} else {
		multiCtx, cancel = context.WithCancel(ctx)
	}

	if config.PoolMaxIterations == nil {
		maxInt := math.MaxInt
		config.PoolMaxIterations = &maxInt
	}

	poolInfo := PoolInfo{WorkerNum: config.WorkerNum}
	timeStart := time.Now()

	defer cancel()

	padding := int(header.Counter)

	wg := sync.WaitGroup{}
	wg.Add(config.WorkerNum)
	ch := make(chan Header, 1)

	go func() {
		wg.Wait()
		close(ch)
	}()

	for i := 0; i < config.WorkerNum; i++ {
		i := i
		go func() {
			defer wg.Done()

			chunkSize := *config.PoolMaxIterations / config.WorkerNum

			sincePos := padding + i*chunkSize
			if i > 0 {
				sincePos += i
			}

			untilPos := sincePos + chunkSize
			if untilPos > *config.PoolMaxIterations {
				untilPos = *config.PoolMaxIterations
			}

			chunkHeader := header
			chunkHeader.Counter = uint64(sincePos)

			calculate, err := Compute(multiCtx, chunkHeader, untilPos)
			if err != nil {
				return
			}

			ch <- calculate
		}()
	}

	for v := range ch {
		if v.IsValid() {
			poolInfo.Header = v
			poolInfo.Time = time.Since(timeStart)
			return poolInfo, nil
		}
	}

	select {
	case <-multiCtx.Done():
		return poolInfo, multiCtx.Err()
	default:
	}

	return poolInfo, ErrMaxIterations
}

// Parse converts the text value of the header into a Header structure
func Parse(header string) (Header, error) {
	var h Header
	tokens := strings.Split(header, ":")

	if len(tokens) < 7 {
		return h, ErrHeaderInvalid
	}

	version, err := strconv.Atoi(tokens[0])
	if err != nil {
		return h, fmt.Errorf("strconv.Atoi: %w", ErrHeaderInvalid)
	}

	difficult, err := strconv.Atoi(tokens[1])
	if err != nil {
		return h, fmt.Errorf("strconv.Atoi: %w", ErrHeaderInvalid)
	}

	expiredAtUnix, err := strconv.ParseInt(tokens[2], 10, 64)
	if err != nil {
		return h, fmt.Errorf("strconv.Atoi: %w", ErrHeaderInvalid)
	}

	subject := tokens[3]
	alg := tokens[4]
	nonce := tokens[5]

	counterByt, err := base64.StdEncoding.DecodeString(tokens[6])
	if err != nil {
		return h, fmt.Errorf("base64.StdEncoding.DecodeString: %w", ErrHeaderInvalid)
	}

	counter := binary.LittleEndian.Uint64(counterByt)

	return Header{
		Version:   version,
		Difficult: difficult,
		Subject:   subject,
		ExpiredAt: expiredAtUnix,
		Alg:       alg,
		Nonce:     nonce,
		Counter:   counter,
	}, nil
}

// Verify checks if the hash corresponds to the task with the given complexity
func Verify(hash string, difficult int) bool {
	if difficult > len(hash) {
		return false
	}

	for idx := range hash[:difficult] {
		if hash[idx] != 0x30 {
			return false
		}
	}

	return true
}

func hashFunc(alg string) hash.Hash {
	var f hash.Hash
	switch alg {
	case AlgSHA256:
		f = sha256.New()
	case AlgSHA512:
		f = sha512.New()
	default:
		return sha1.New()
	}

	return f
}
