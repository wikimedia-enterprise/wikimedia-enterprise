// Package creates HTTP handler for getuser endpoint.
package getuser

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"wikimedia-enterprise/api/auth/config/env"
	"wikimedia-enterprise/api/auth/submodules/httputil"
	"wikimedia-enterprise/api/auth/submodules/log"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"go.uber.org/dig"
)

// Parameters dependency injection for the handler.
type Parameters struct {
	dig.In
	Redis redis.Cmdable
	Env   *env.Environment
}

// Response structure represents response data format.
type Response struct {
	Groups                []string         `json:"groups"`
	OndemandRequestsCount int              `json:"ondemand_requests_count"`
	OndemandLimit         int              `json:"ondemand_limit"`
	SnapshotRequestsCount int              `json:"snapshot_requests_count"`
	SnapshotLimit         int              `json:"snapshot_limit"`
	ChunkRequestsCount    int              `json:"chunk_requests_count"`
	ChunkLimit            int              `json:"chunk_limit"`
	Username              string           `json:"username"`
	Apis                  []env.AccessPath `json:"apis"`
}

var (
	// Here, "internal" means the error is on our side (Wikimedia Enterprise), not necessarily in the auth API server.
	internalErr = errors.New("Internal error, please try again later.")
)

// NewHandler creates new getuser HTTP handler.
func NewHandler(p *Parameters) gin.HandlerFunc {
	return func(gcx *gin.Context) {
		mur, exists := gcx.Get("user")

		if !exists {
			log.Error("problem getting user")
			httputil.InternalServerError(gcx, errors.New("User does not exist!"))
			return
		}

		usr, ok := mur.(*httputil.User)

		if !ok {
			log.Error("problem getting user - unknown type of user identity")
			httputil.InternalServerError(gcx, errors.New("Unknown type of user identity!"))
			return
		}

		if len(usr.GetGroups()) == 0 {
			log.Error("problem getting user - User has no allocated permissions")
			httputil.InternalServerError(gcx, errors.New("User has no allocated permissions!"))
			return
		}

		ondemandCount, err := p.Redis.Get(
			gcx,
			fmt.Sprintf("cap:ondemand:user:%s:count", usr.Username),
		).Int()

		if err == redis.Nil {
			ondemandCount = 0
		} else if err != nil {
			log.Error(err, log.Tip("problem getting ondemand count"))
			httputil.InternalServerError(gcx, internalErr)
			return
		}

		snapshotCount, err := p.Redis.Get(
			gcx,
			fmt.Sprintf("cap:snapshot:user:%s:count", usr.Username),
		).Int()

		if err == redis.Nil {
			snapshotCount = 0
		} else if err != nil {
			log.Error(err, log.Tip("problem getting snapshot count"))
			httputil.InternalServerError(gcx, internalErr)
			return
		}

		chunkReqCount, err := p.Redis.Get(
			gcx,
			fmt.Sprintf("cap:chunk:user:%s:count", usr.Username),
		).Int()

		if err == redis.Nil {
			chunkReqCount = 0
		} else if err != nil {
			log.Error(err, log.Tip("problem getting chunks count"))
			httputil.InternalServerError(gcx, internalErr)
			return
		}

		gps := usr.GetGroups()
		var aps = []env.AccessPath{}

		for _, grp := range gps {
			for gp, pts := range p.Env.AccessPolicy.Map {
				if grp == gp {
					aps = append(aps, pts...)
				}
			}
		}

		if len(p.Env.AccessPolicy.Map["*"]) != 0 {
			aps = append(aps, p.Env.AccessPolicy.Map["*"]...)
		}

		ondemandLimit, _ := strconv.Atoi(p.Env.OndemandLimit)
		snapshotLimit, _ := strconv.Atoi(p.Env.SnapshotLimit)
		chunkLimit, _ := strconv.Atoi(p.Env.ChunkLimit)

		r := new(Response)
		r.Username = usr.Username
		r.Groups = gps
		r.OndemandRequestsCount = ondemandCount
		r.SnapshotRequestsCount = snapshotCount
		r.ChunkRequestsCount = chunkReqCount
		r.OndemandLimit = ondemandLimit
		r.SnapshotLimit = snapshotLimit
		r.ChunkLimit = chunkLimit
		r.Apis = aps

		gcx.JSON(http.StatusOK, r)
	}
}
