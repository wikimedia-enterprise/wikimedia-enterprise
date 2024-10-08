// Package creates HTTP handler for getuser endpoint.
package getuser

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"wikimedia-enterprise/api/auth/config/env"
	"wikimedia-enterprise/general/httputil"
	"wikimedia-enterprise/general/log"

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
	Username              string           `json:"username"`
	Apis                  []env.AccessPath `json:"apis"`
}

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

		rdc, err := p.Redis.Get(
			gcx,
			fmt.Sprintf("cap:ondemand:user:%s:count", usr.Username),
		).Int()

		if err == redis.Nil {
			rdc = 0
		} else if err != nil {
			log.Error(err, log.Tip("problem getting user from redis"))
			httputil.InternalServerError(gcx, err)
			return
		}

		rsc, err := p.Redis.Get(
			gcx,
			fmt.Sprintf("cap:snapshot:user:%s:count", usr.Username),
		).Int()

		if err == redis.Nil {
			rsc = 0
		} else if err != nil {
			log.Error(err, log.Tip("problem getting user from redis"))
			httputil.InternalServerError(gcx, err)
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

		odl, _ := strconv.Atoi(p.Env.OndemandLimit)
		snl, _ := strconv.Atoi(p.Env.SnapshotLimit)

		r := new(Response)
		r.Username = usr.Username
		r.Groups = gps
		r.OndemandRequestsCount = rdc
		r.SnapshotRequestsCount = rsc
		r.OndemandLimit = odl
		r.SnapshotLimit = snl
		r.Apis = aps

		gcx.JSON(http.StatusOK, r)
	}
}
