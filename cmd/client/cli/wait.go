package cli

import (
	"context"
	"log"
	"time"

	proto "gitlab.com/netbook-devs/spawner-service/proto/netbookai/spawner"
)

//waitUntilClusterReady Wait until the give cluster return "ACTIVE" status
//
// This will run for about 12min, with multiuple iteration of sleeps.
// Starting with 6min, then 4min, 2min.
// Cluster can become fully active in any one of sleep cycle, as soon as we see status is ACTIVE, we return it.
// When sleep counter reaches 0, we return FAILED status, 15min is too long for cluster to become active
func waitUntilClusterReady(ctx context.Context, c proto.SpawnerServiceClient, req *proto.ClusterRequest) string {
	clusterStatus := &proto.ClusterStatusRequest{
		Provider:    req.Provider,
		Region:      req.Region,
		ClusterName: req.ClusterName,
	}

	waitfor := 6 //minute

	for {
		log.Println("wait for cluster to be active, duration ", waitfor, "min")
		time.Sleep(time.Duration(waitfor) * time.Minute)

		//check cluster status and if active return
		stat, err := c.ClusterStatus(ctx, clusterStatus)
		if err != nil {
			log.Println("failed to fetch cluster stat")
		}

		if stat.Status == "ACTIVE" {
			return stat.Status
		}
		waitfor = waitfor - 2

		if waitfor == 0 {
			//already waited for 12+min
			return "FAILED"
		}
	}
}
