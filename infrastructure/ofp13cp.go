package infrastructure

import (
	// "arieoldman/arieoldman/krios/common"
	"github.com/netrack/openflow/ofp"
	// "github.com/netrack/openflow/ofputil"
	of "github.com/netrack/openflow"
	"github.com/golang/glog"
	"fmt"
	"arieoldman/arieoldman/krios/controller"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"bytes"
)

// OpenFlow13ControlPlane is an OpenFlow 1.3 control plane.
type OpenFlow13ControlPlane struct {
	ctrlSession controller.SessionManager
}

// Start will start the control plane listener
func (cp *OpenFlow13ControlPlane) Start(port uint16) {
	echoRequestEvent := of.TypeMatcher(of.TypeEchoRequest)
	featuresReplyEvent := of.TypeMatcher(of.TypeFeaturesReply)
	helloEvent := of.TypeMatcher(of.TypeHello)
	errorEvent := of.TypeMatcher(of.TypeError)

	mux := of.NewServeMux()

	mux.HandleFunc(errorEvent, func(rw of.ResponseWriter, r *of.Request){
		var packet ofp.Error
		packet.ReadFrom(r.Body)

		glog.Errorln("Error:",packet.Error())
	})

	mux.HandleFunc(featuresReplyEvent, func(rw of.ResponseWriter, r * of.Request){
		var featuresReply ofp.SwitchFeatures
		featuresReply.ReadFrom(r.Body)

		glog.Infof("Features Reply from %s: DatapathID %x, %v\n",
			r.Addr, featuresReply.DatapathID, featuresReply)

		if can := cp.ctrlSession.CanHandshake(string(r.Addr)); can == true {

		}
	})

	mux.HandleFunc(helloEvent, func(rw of.ResponseWriter, r *of.Request){
		//Send back the Hello response

		glog.Infoln("Responded to", of.TypeHello, "from host", r.Addr, ".")

		rw.Write(&of.Header{Type: of.TypeHello}, nil)

		// Features Request
		glog.Infoln("Features Request to ", r.Addr)
		rw.Write(&of.Header{Type: of.TypeFeaturesRequest}, nil)

	})

	mux.HandleFunc(echoRequestEvent, func( rw of.ResponseWriter, r *of.Request){
		glog.Infoln("Echo request from",r.Addr,". Replying.")

		var req ofp.EchoRequest
		req.ReadFrom(r.Body)

		echoReply := of.NewRequest(of.TypeEchoReply, &ofp.EchoReply{
			Data: req.Data,
		})

		rw.Write(&of.Header{Type: of.TypeEchoReply}, echoReply)
	})

	glog.Info("Control plane firing up engines.")
	
	of.ListenAndServe(fmt.Sprintf(":%d",port), mux)
}

// Stop will kill the control plane listener
func (cp *OpenFlow13ControlPlane) Stop() {

}