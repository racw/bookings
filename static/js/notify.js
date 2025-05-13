function notify(msg, type) {
    notie.alert({
        type: type,
        text: msg,
    })
}

function notifyModal(title, text, icon, confirmationButtonText) {
    Swal.fire({
        title: title,
        text: text,
        icon: icon,
        confirmButtonText: confirmationButtonText,
    })
}
