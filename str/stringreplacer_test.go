package main

import (
	"fmt"
	"strings"
	"strconv"
	//"time"
	"testing"
)
var (
	adcode = `<!DOCTYPE html>
		<HTML>
		
		<HEAD></HEAD>
		
		<BODY>
			<script type="text/javascript">
				function callNative(command) {
					console.log("callNative called");
					var iframe = document.createElement("IFRAME");
					iframe.setAttribute("src", "mnet://" + command);
					document.documentElement.appendChild(iframe);
					iframe.parentNode.removeChild(iframe);
					iframe = null;
				}
				window.mnMobileAdLoaded = function(width, height) {
					console.log('ad loaded')
					callNative('ybnca/adloaded?width=' + width + '&height=' + height)
				}
		
				window.mnMobileAdFailed = function() {
					console.log('ad failed')
					callNative('ybnca/adfailed')
				}
			</script>
			<script type="text/javascript">
				window._mNHandle = window._mNHandle || {};
				window._mNHandle.queue = window._mNHandle.queue || [];
				medianet_versionId = '111299';
				medianet_width = '${WIDTH}';
				medianet_height = '${HEIGHT}';
				medianet_hint = '${HINT}';
				medianet_requrl = '${REQ_URL}';
				medianet_chnm2 = '${CID}';
				medianet_chnm3 = '${CRID}';
				medianet_bdrId = '${BIDDER_ID}';
				medianet_sbdrId = '';
				medianet_auctionid = '${ACID}';
				medianet_misc = {};
				medianet_misc.matchString = 'null';
			</script>
			<script src="http://contextual.media.net/dmedianet.js?cid=${ECID}" async="async"></script>
			<div id="432114897">
				<script type="text/javascript">
					try {
						window._mNHandle.queue.push(function() {
							window._mNDetails.loadTag("${EPC}", "${SIZE}", "${EPC}");
						});
					} catch (error) {}
				</script>
			</div>
		</BODY>
		
		</HTML>`
	
	adcode1 = `<!DOCTYPE html>
		<HTML>
		
		<HEAD></HEAD>
		
		<BODY>
			<script type="text/javascript">
				function callNative(command) {
					console.log("callNative called");
					var iframe = document.createElement("IFRAME");
					iframe.setAttribute("src", "mnet://" + command);
					document.documentElement.appendChild(iframe);
					iframe.parentNode.removeChild(iframe);
					iframe = null;
				}
				window.mnMobileAdLoaded = function(width, height) {
					console.log('ad loaded')
					callNative('ybnca/adloaded?width=' + width + '&height=' + height)
				}
		
				window.mnMobileAdFailed = function() {
					console.log('ad failed')
					callNative('ybnca/adfailed')
				}
			</script>
			<script type="text/javascript">
				window._mNHandle = window._mNHandle || {};
				window._mNHandle.queue = window._mNHandle.queue || [];
				medianet_versionId = '111299';
				medianet_width = '%s';
				medianet_height = '%s';
				medianet_hint = '%s';
				medianet_requrl = '%s';
				medianet_chnm2 = '%s';
				medianet_chnm3 = '%s';
				medianet_bdrId = '%s';
				medianet_sbdrId = '';
				medianet_auctionid = '%s';
				medianet_misc = {};
				medianet_misc.matchString = 'null';
			</script>
			<script src="http://contextual.media.net/dmedianet.js?cid=%s" async="async"></script>
			<div id="432114897">
				<script type="text/javascript">
					try {
						window._mNHandle.queue.push(function() {
							window._mNDetails.loadTag("%s", "%s", "%s");
						});
					} catch (error) {}
				</script>
			</div>
		</BODY>
		
		</HTML>`
	macros = []string{"${WIDTH}", "${HEIGHT}", "${HINT}","${REQ_URL}","${CID}","${CRID}","${BIDDER_ID}","${ACID}","${ECID}","${EPC}","${SIZE}"}
)


func BenchmarkReplacer(b *testing.B)  {
	for i:=0; i<b.N; i++ {
		r := strings.NewReplacer(macros[0], strconv.Itoa(300),
			macros[1], strconv.Itoa(300),
			macros[2], "fhdjnvmdfdvdfvdvdfv/bhvfdbvhdfvhfdim",
			macros[3], "vfdvjndfvjfdvfdvdfvm,.fdjnvjfdnvjfdvfjd",
			macros[4], "publisher123",
			macros[5], "cr123",
			macros[6], strconv.Itoa(3000),
			macros[7], "cdcdcddcdcdcdcdcdcdcd",
			macros[8], "ec1212121",
			macros[9], "epc12121",
			macros[10], "300x250")
		r.Replace(adcode)
	}
}

func BenchmarkReplace(b *testing.B)  {
	for i:=0; i<b.N; i++ {
		ma1 := strings.Replace(adcode, macros[0], strconv.Itoa(300), -1)
		ma1 = strings.Replace(ma1, macros[1], strconv.Itoa(300), -1)
		ma1 = strings.Replace(ma1, macros[2], "fhdjnvmdfdvdfvdvdfv/bhvfdbvhdfvhfdim", -1)
		ma1 = strings.Replace(ma1, macros[3], "vfdvjndfvjfdvfdvdfvm,.fdjnvjfdnvjfdvfjd", -1)
		ma1 = strings.Replace(ma1, macros[4], "publisher123", -1)
		ma1 = strings.Replace(ma1, macros[5], "cr123", -1)
		ma1 = strings.Replace(ma1, macros[6], strconv.Itoa(3000), -1)
		ma1 = strings.Replace(ma1, macros[7], "cdcdcddcdcdcdcdcdcdcd", -1)
		ma1 = strings.Replace(ma1, macros[8], "ec1212121", -1)
		ma1 = strings.Replace(ma1, macros[9], "epc12121", -1)
		ma1 = strings.Replace(ma1, macros[10], "300x250", -1)
	}
}

func BenchmarkPrintf(b *testing.B) {
	for i:=0; i<100000; i++ {
		fmt.Sprint(adcode1,
			strconv.Itoa(300),
			strconv.Itoa(300),
			"fhdjnvmdfdvdfvdvdfv/bhvfdbvhdfvhfdim",
			"vfdvjndfvjfdvfdvdfvm,.fdjnvjfdnvjfdvfjd",
			"publisher123",
			"cr123",
			strconv.Itoa(3000),
			"cdcdcddcdcdcdcdcdcdcd",
			"ec1212121",
			"epc12121",
			"300x250",
			"ec1212121")
	}
}
//func main() {
//	t:= time.Now()
//	var ma string
//	var ma1 string
//	var ma2 string
//	for i:=0; i<100000; i++ {
//		r := strings.NewReplacer(macros[0], strconv.Itoa(300),
//			macros[1], strconv.Itoa(300),
//			macros[2], "fhdjnvmdfdvdfvdvdfv/bhvfdbvhdfvhfdim",
//			macros[3], "vfdvjndfvjfdvfdvdfvm,.fdjnvjfdnvjfdvfjd",
//			macros[4], "publisher123",
//			macros[5], "cr123",
//			macros[6], strconv.Itoa(3000),
//			macros[7], "cdcdcddcdcdcdcdcdcdcd",
//			macros[8], "ec1212121",
//			macros[9], "epc12121",
//			macros[10], "300x250")
//		ma = r.Replace(adcode)
//	}
//	fmt.Println(ma, "\n",time.Since(t),time.Since(t)/time.Duration(100000))
//
//	t1:=time.Now()
//	for i:=0; i<100000; i++ {
//		ma1 = strings.Replace(adcode, macros[0], strconv.Itoa(300), -1)
//		ma1 = strings.Replace(ma1, macros[1], strconv.Itoa(300), -1)
//		ma1 = strings.Replace(ma1, macros[2], "fhdjnvmdfdvdfvdvdfv/bhvfdbvhdfvhfdim", -1)
//		ma1 = strings.Replace(ma1, macros[3], "vfdvjndfvjfdvfdvdfvm,.fdjnvjfdnvjfdvfjd", -1)
//		ma1 = strings.Replace(ma1, macros[4], "publisher123", -1)
//		ma1 = strings.Replace(ma1, macros[5], "cr123", -1)
//		ma1 = strings.Replace(ma1, macros[6], strconv.Itoa(3000), -1)
//		ma1 = strings.Replace(ma1, macros[7], "cdcdcddcdcdcdcdcdcdcd", -1)
//		ma1 = strings.Replace(ma1, macros[8], "ec1212121", -1)
//		ma1 = strings.Replace(ma1, macros[9], "epc12121", -1)
//		ma1 = strings.Replace(ma1, macros[10], "300x250", -1)
//	}
//	fmt.Println(ma1, "\n",time.Since(t),time.Since(t1)/time.Duration(100000))
//
//	t2:=time.Now()
//	for i:=0; i<100000; i++ {
//		ma2 = fmt.Sprint(adcode1,
//			strconv.Itoa(300),
//			strconv.Itoa(300),
//			"fhdjnvmdfdvdfvdvdfv/bhvfdbvhdfvhfdim",
//			"vfdvjndfvjfdvfdvdfvm,.fdjnvjfdnvjfdvfjd",
//			"publisher123",
//			"cr123",
//			strconv.Itoa(3000),
//			"cdcdcddcdcdcdcdcdcdcd",
//			"ec1212121",
//			"epc12121",
//			"300x250",
//			"ec1212121")
//	}
//	fmt.Println(ma2, "\n",time.Since(t2),time.Since(t2)/time.Duration(100000))
//}