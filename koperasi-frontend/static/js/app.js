// ===== Auto-dismiss flash alerts after 4 seconds =====
document.addEventListener("DOMContentLoaded", function () {
    document.querySelectorAll(".alert-dismissible.auto-dismiss").forEach(function (el) {
        setTimeout(function () {
            const alert = bootstrap.Alert.getOrCreateInstance(el);
            alert.close();
        }, 4000);
    });

    // ===== Active nav pill auto-scroll into view (mobile) =====
    var activeNav = document.querySelector(".nav-pills .nav-link.active");
    if (activeNav) {
        activeNav.scrollIntoView({ behavior: "smooth", inline: "center", block: "nearest" });
    }

    // ===== Confirm before destructive actions =====
    document.querySelectorAll("[data-confirm]").forEach(function (el) {
        el.addEventListener("click", function (e) {
            if (!confirm(el.getAttribute("data-confirm"))) {
                e.preventDefault();
            }
        });
    });

    // ===== Tooltip init for all elements with title =====
    var tooltipTriggerList = [].slice.call(document.querySelectorAll('[data-bs-toggle="tooltip"]'));
    tooltipTriggerList.forEach(function (el) {
        new bootstrap.Tooltip(el);
    });

    // ===== Cart quantity buttons =====
    document.querySelectorAll(".btn-qty-minus").forEach(function (btn) {
        btn.addEventListener("click", function (e) {
            e.preventDefault();
            var form = btn.closest("form");
            var input = form.querySelector('input[name="jumlah"]');
            var min = parseInt(btn.getAttribute("data-min")) || 1;
            var val = Math.max(min, parseInt(input.value) - 1);
            input.value = val;
            form.submit();
        });
    });
    document.querySelectorAll(".btn-qty-plus").forEach(function (btn) {
        btn.addEventListener("click", function (e) {
            e.preventDefault();
            var form = btn.closest("form");
            var input = form.querySelector('input[name="jumlah"]');
            var max = parseInt(btn.getAttribute("data-max")) || 999;
            var val = Math.min(max, parseInt(input.value) + 1);
            input.value = val;
            form.submit();
        });
    });
});
