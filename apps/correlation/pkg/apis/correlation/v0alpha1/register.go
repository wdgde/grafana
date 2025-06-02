package v0alpha1

import (
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/grafana/grafana/pkg/apimachinery/utils"
)

const (
	GROUP      = "correlation.grafana.app"
	VERSION    = "v0alpha1"
	APIVERSION = GROUP + "/" + VERSION

	// Resource constants
	CORRELATION_RESOURCE = "correlations"
)

var CorrelationResourceInfo = utils.NewResourceInfo(GROUP, VERSION,
	"correlations", "correlation", "Correlation",
	func() runtime.Object { return &Correlation{} },
	func() runtime.Object { return &CorrelationList{} },
	utils.TableColumns{
		Definition: []metav1.TableColumnDefinition{
			{Name: "Name", Type: "string", Format: "name"},
			{Name: "Title", Type: "string", Format: "string", Description: "The dashboard name"},
			{Name: "Created At", Type: "date"},
		},
		Reader: func(obj any) ([]interface{}, error) {
			c, ok := obj.(*Correlation)
			if ok {
				if c != nil {
					return []interface{}{
						c.Name,
						c.CreationTimestamp.UTC().Format(time.RFC3339),
					}, nil
				}
			}
			return nil, fmt.Errorf("expected correlation")
		},
	},
)
