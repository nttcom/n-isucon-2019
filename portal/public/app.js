///// ベンチ用　/////

const benchButton = document.getElementById('bench');
if (benchButton) {
    benchButton.addEventListener('click', async () => {
        try {
            const res = await axios.get('request_bench');
            console.log(res.data);
        } catch (e) {
            console.error(e);
        }
        $('#bench_accept').fadeIn(1000).delay(7000).fadeOut(2000);
    });
}
