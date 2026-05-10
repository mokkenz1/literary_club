// Функция для раскрытия карточек мероприятий
function toggleEvent(card) {
    // Закрываем все другие карточки
    document.querySelectorAll('.event-card').forEach(c => {
        if (c !== card && c.classList.contains('expanded')) {
            c.classList.remove('expanded');
        }
    });
    
    // Переключаем текущую
    card.classList.toggle('expanded');
}

// Анимация появления
document.addEventListener('DOMContentLoaded', function() {
    const fadeElements = document.querySelectorAll('.fade-up');
    
    const observer = new IntersectionObserver((entries) => {
        entries.forEach(entry => {
            if (entry.isIntersecting) {
                entry.target.classList.add('visible');
                observer.unobserve(entry.target);
            }
        });
    }, {
        threshold: 0.1,
        rootMargin: '0px 0px -30px 0px'
    });

    fadeElements.forEach(el => observer.observe(el));

    // Закрываем карточки при клике вне
    document.addEventListener('click', function(e) {
        if (!e.target.closest('.event-card')) {
            document.querySelectorAll('.event-card.expanded').forEach(card => {
                card.classList.remove('expanded');
            });
        }
    });
});