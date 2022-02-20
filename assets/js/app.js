function parseScanContents(scannedContent) {
    const cTrimmed = scannedContent.trim()

    // Return early if the scanned code is a barcode
    if(scannedContent.length<20){
        return (!isNaN(parseInt(cTrimmed)))?{type:"ean-13", content:cTrimmed}:{type:"error", content:undefined}
    }

    if(!cTrimmed.includes("https://") || !cTrimmed.includes("?d=")){
        return {type:"error", content:undefined}
    }

    let encoded_shit = cTrimmed.substr(cTrimmed.lastIndexOf("?d=")+3)
    let ticket_type = cTrimmed.split("/")[4]
    try {
        encoded_shit = decodeURIComponent(encoded_shit)
        ticket_type = parseInt(ticket_type)
      } catch(e) {
        return {type:"error", content:undefined}
      }
    return {type:"qr", content:encoded_shit, ticket_type_id:ticket_type}
}


function listenerSetup(){
    const area = document.getElementById("scanarea")
    area.addEventListener('keyup', (e) => {
        if (e.key === 'Enter') {
            console.log(parseScanContents(area.value))
            area.value = ""

        }
      });
}