package core

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/kr/pretty"

	"github.com/ohsu-comp-bio/funnel/tes"
	"github.com/ohsu-comp-bio/funnel/tests"
	"github.com/prometheus/common/expfmt"
)

func TestMetrics(t *testing.T) {
	conf := tests.DefaultConfig()
	if conf.Database != "mongodb" {
		t.Skip("metrics are only supported by mongodb")
	}

	fun := tests.NewFunnel(conf)
	fun.StartServer()
	tests.SetLogOutput(log, t)

	id1 := fun.Run(`'echo hello world'`)
	id2 := fun.Run(`'echo hello world'`)
	id3 := fun.Run(`'exit 1'`)
	id4 := fun.Run(`'sleep 100'`)
	_ = id4

	fun.Wait(id1)
	fun.Wait(id2)
	fun.Wait(id3)

	bg := context.Background()
	resp, err := fun.HTTP.GetServiceInfo(bg, &tes.ServiceInfoRequest{})
	if err != nil {
		t.Fatal(err)
	}

	log.Info("INFO", resp)
	counts := resp.TaskStateCounts
	if len(counts) != 9 || counts["COMPLETE"] != 2 || counts["EXECUTOR_ERROR"] != 1 || counts["RUNNING"] != 1 {
		t.Error("unexpected counts from service info")
	}

	hresp, err := http.Get(fun.Conf.Server.HTTPAddress() + "/metrics")
	if err != nil {
		t.Fatal(err)
	}
	defer hresp.Body.Close()

	parser := expfmt.TextParser{}
	met, err := parser.TextToMetricFamilies(hresp.Body)
	if err != nil {
		t.Fatal(err)
	}

	// TODO unfortunately, we have to wait for the prometheus poller to update
	time.Sleep(20 * time.Second)

	// It's tedious to check these so just find one.
	for _, m := range met["funnel_tasks_state_count"].Metric {
		if *m.Label[0].Name == "state" && *m.Label[0].Value == "COMPLETE" {
			if *m.Gauge.Value != 2 {
				t.Error("unexpected counts from prometheus endpoint", *m.Gauge.Value)
			}
		}
	}

	pretty.Println(met["funnel_tasks_state_count"].Metric)
}
