export const formatDate = (date: Date) => {
    return `${('0' + date.getDate()).slice(-2)}/${('0' + String(date.getMonth() + 1)).slice(-2)}/${date.getFullYear()} ${('0' + date.getHours()).slice(-2)}:${('0' + date.getMinutes()).slice(-2)}:${('0' + date.getSeconds()).slice(-2)}`
}

export const TimeToString = (seconds: number) => {
    const times: string[] = []
    if (Math.floor(seconds / 604800) > 0) {
        times.push(`${Math.floor(seconds / 604800)} week${Math.floor(seconds / 604800) != 1 ? "s" : ""}`)
        seconds %= 604800;
    }
    if (Math.floor(seconds / 86400) > 0) {
        times.push(`${Math.floor(seconds / 86400)} day${Math.floor(seconds / 86400) != 1 ? "s" : ""}`)
        seconds %= 86400;
    }
    if (Math.floor(seconds / 3600) > 0) {
        times.push(`${Math.floor(seconds / 3600)} hour${Math.floor(seconds / 3600) != 1 ? "s" : ""}`)
        seconds %= 3600;
    }
    if (Math.floor(seconds / 60) > 0) {
        times.push(`${Math.floor(seconds / 60)} minute${Math.floor(seconds / 60) != 1 ? "s" : ""}`)
        seconds %= 60;
    }
    if (seconds > 0) {
        times.push(`${Math.floor(seconds)} second${Math.floor(seconds) != 1 ? "s" : ""}`)
    }

    const last = times.pop();

    return times.length == 0 ? last : `${times.join(", ")} and ${last}`;
}