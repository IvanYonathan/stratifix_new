document.addEventListener('DOMContentLoaded', function() {
    // DOM Elements
    const loginSection = document.getElementById('login-section');
    const adminDashboard = document.getElementById('admin-dashboard');
    const loginForm = document.getElementById('login-form');
    const bookingsTableBody = document.getElementById('bookings-table-body');
    const noBookingsMessage = document.getElementById('no-bookings');
    const refreshButton = document.getElementById('refresh-bookings');
    const logoutButton = document.getElementById('logout');

    // Check if already authenticated
    function checkAuth() {
        const isAuthenticated = localStorage.getItem('admin-authenticated');
        if (isAuthenticated === 'true') {
            showDashboard();
            fetchBookings();
        }
    }

    // Handle login
    function handleLogin(e) {
        e.preventDefault();
        
        const username = document.getElementById('username').value;
        const password = document.getElementById('password').value;
        
        if (!username || !password) {
            alert('Please enter username and password.');
            return;
        }
        
        // Send login request to the server
        fetch('/api/admin/login', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                username: username,
                password: password
            }),
        })
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                localStorage.setItem('admin-authenticated', 'true');
                showDashboard();
                fetchBookings();
            } else {
                alert('Invalid credentials.');
            }
        })
        .catch(error => {
            console.error('Error:', error);
            alert('Login failed. Please try again.');
        });
    }

    // Show dashboard, hide login
    function showDashboard() {
        loginSection.style.display = 'none';
        adminDashboard.style.display = 'block';
    }

    // Fetch bookings from the server
    function fetchBookings() {
        fetch('/api/admin/bookings')
            .then(response => response.json())
            .then(data => {
                if (data.success && data.data) {
                    renderBookings(data.data);
                } else {
                    renderBookings([]);
                }
            })
            .catch(error => {
                console.error('Error fetching bookings:', error);
                renderBookings([]);
            });
    }

    // Render bookings in the table
    function renderBookings(bookings) {
        bookingsTableBody.innerHTML = '';
        
        if (bookings.length === 0) {
            noBookingsMessage.style.display = 'block';
            return;
        }
        
        noBookingsMessage.style.display = 'none';
        
        bookings.forEach(booking => {
            const tr = document.createElement('tr');
            
            // Format date
            const date = new Date(booking.createdAt);
            const formattedDate = `${date.toLocaleDateString()} ${date.toLocaleTimeString()}`;
            
            tr.innerHTML = `
                <td>${booking.id}</td>
                <td>${booking.customerName}</td>
                <td>
                    ${booking.customerEmail}<br>
                    ${booking.customerPhone}
                </td>
                <td>${booking.seats.join(', ')}</td>
                <td>${booking.paymentMethod === 'bank_transfer' ? 'Bank Transfer' : 'Cash'}</td>
                <td>${formatPrice(booking.totalAmount)}</td>
                <td>${formattedDate}</td>
                <td class="action-buttons">
                    <button class="confirm-btn" data-id="${booking.id}">Confirm</button>
                    <button class="cancel-btn" data-id="${booking.id}">Cancel</button>
                </td>
            `;
            
            bookingsTableBody.appendChild(tr);
        });
        
        // Add event listeners to buttons
        document.querySelectorAll('.confirm-btn').forEach(button => {
            button.addEventListener('click', (e) => {
                const bookingId = parseInt(e.target.dataset.id);
                handleConfirmPayment(bookingId);
            });
        });
        
        document.querySelectorAll('.cancel-btn').forEach(button => {
            button.addEventListener('click', (e) => {
                const bookingId = parseInt(e.target.dataset.id);
                handleCancelBooking(bookingId);
            });
        });
    }

    // Handle confirming payment
    function handleConfirmPayment(bookingId) {
        if (confirm('Are you sure you want to confirm this payment?')) {
            // Send confirmation to server
            fetch('/api/admin/verify', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    bookingId: bookingId,
                    status: 'confirmed'
                }),
            })
            .then(response => response.json())
            .then(data => {
                if (data.success) {
                    alert(`Payment for booking #${bookingId} confirmed!`);
                    fetchBookings(); // Refresh the bookings list
                } else {
                    alert(`Error: ${data.message}`);
                }
            })
            .catch(error => {
                console.error('Error:', error);
                alert('Failed to confirm payment. Please try again.');
            });
        }
    }

    // Handle canceling booking
    function handleCancelBooking(bookingId) {
        if (confirm('Are you sure you want to cancel this booking?')) {
            // Send cancellation to server
            fetch('/api/admin/verify', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    bookingId: bookingId,
                    status: 'cancelled'
                }),
            })
            .then(response => response.json())
            .then(data => {
                if (data.success) {
                    alert(`Booking #${bookingId} cancelled!`);
                    fetchBookings(); // Refresh the bookings list
                } else {
                    alert(`Error: ${data.message}`);
                }
            })
            .catch(error => {
                console.error('Error:', error);
                alert('Failed to cancel booking. Please try again.');
            });
        }
    }

    // Handle logout
    function handleLogout() {
        localStorage.removeItem('admin-authenticated');
        loginSection.style.display = 'block';
        adminDashboard.style.display = 'none';
        loginForm.reset();
    }

    // Format price as IDR
    function formatPrice(price) {
        return new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR' }).format(price);
    }

    // Initialize the app
    function init() {
        checkAuth();
        
        // Event listeners
        loginForm.addEventListener('submit', handleLogin);
        refreshButton.addEventListener('click', fetchBookings);
        logoutButton.addEventListener('click', handleLogout);
    }

    // Start the application
    init();
});