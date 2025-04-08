document.addEventListener('DOMContentLoaded', function() {
    // Constants for pricing
    const PRICES = {
        'regular': 25000,
        'vip': 30000
    };

    // State management
    const state = {
        seats: [],
        selectedSeats: [],
        totalAmount: 0
    };

    // DOM Elements
    const seatingMap = document.querySelector('.seating-map');
    const selectedSeatsList = document.getElementById('selected-seats-list');
    const totalAmountElement = document.getElementById('total-amount');
    const seatInput = document.getElementById('seat-input');
    const enterSeatsButton = document.getElementById('enter-seats');
    const customerForm = document.getElementById('customer-form');
    const bookingForm = document.getElementById('booking-form');
    const bookingConfirmation = document.getElementById('booking-confirmation');
    const newBookingButton = document.getElementById('new-booking');
    const paymentMethodRadios = document.getElementsByName('payment-method');
    const bankTransferInfo = document.querySelector('.bank-transfer-info');
    const cashInfo = document.querySelector('.cash-info');

    // Fetch all seats from the server
    function fetchSeats() {
        fetch('/api/seats')
            .then(response => response.json())
            .then(data => {
                if (data.success && data.data) {
                    state.seats = data.data;
                    renderSeats();
                } else {
                    console.error('Failed to fetch seats:', data.message);
                }
            })
            .catch(error => {
                console.error('Error fetching seats:', error);
            });
    }

    // Render seats in the theater layout
    // Render seats in the theater layout
    function renderSeats() {
        seatingMap.innerHTML = '';
        
        // Group seats by row
        const seatsByRow = {};
        state.seats.forEach(seat => {
            if (!seatsByRow[seat.row]) {
                seatsByRow[seat.row] = [];
            }
            seatsByRow[seat.row].push(seat);
        });
        
        // Sort rows alphabetically
        const sortedRows = Object.keys(seatsByRow).sort();
        
        // Create row elements
        sortedRows.forEach(row => {
            const rowDiv = document.createElement('div');
            rowDiv.className = 'seat-row';
            
            // Get the seats for this row and sort by number
            const rowSeats = seatsByRow[row].sort((a, b) => a.number - b.number);
            
            // Create a map of seat numbers to seat objects for easier lookup
            const seatMap = {};
            rowSeats.forEach(seat => {
                seatMap[seat.number] = seat;
            });
            
            // First add a consistent margin at the start
            const startSpacer = document.createElement('div');
            startSpacer.className = 'seat-spacer';
            startSpacer.style.width = '10px';
            rowDiv.appendChild(startSpacer);
            
            // Left section (seats 1-15)
            for (let i = 1; i <= 15; i++) {
                if (seatMap[i]) {
                    // If seat exists, create a seat element
                    const seatDiv = createSeatElement(seatMap[i]);
                    rowDiv.appendChild(seatDiv);
                } else {
                    // Create a placeholder even if seat doesn't exist for consistent spacing
                    const placeholderDiv = document.createElement('div');
                    placeholderDiv.className = 'seat-placeholder';
                    rowDiv.appendChild(placeholderDiv);
                }
            }
            
            // Row label in the middle
            const rowLabel = document.createElement('div');
            rowLabel.className = 'row-label';
            rowLabel.textContent = row;
            rowDiv.appendChild(rowLabel);
            
            // Right section (seats 16-30)
            for (let i = 16; i <= 30; i++) {
                if (seatMap[i]) {
                    // If seat exists, create a seat element
                    const seatDiv = createSeatElement(seatMap[i]);
                    rowDiv.appendChild(seatDiv);
                } else {
                    // Create a placeholder even if seat doesn't exist for consistent spacing
                    const placeholderDiv = document.createElement('div');
                    placeholderDiv.className = 'seat-placeholder';
                    rowDiv.appendChild(placeholderDiv);
                }
            }
            
            // End spacer for consistency
            const endSpacer = document.createElement('div');
            endSpacer.className = 'seat-spacer';
            endSpacer.style.width = '10px';
            rowDiv.appendChild(endSpacer);
            
            seatingMap.appendChild(rowDiv);
        });
    }
    // Helper function to create a seat element
function createSeatElement(seat) {
    const seatDiv = document.createElement('div');
    
    // Create the appropriate class based on both type and status
    let classNames = `seat ${seat.type}`;
    
    // Add status class
    if (seat.status === 'sponsored') {
        classNames += ' sponsored'; // Add sponsored class for sponsored status
    } else {
        classNames += ` ${seat.status}`; // Add regular status class
    }
    
    seatDiv.className = classNames;
    seatDiv.dataset.id = seat.id;
    seatDiv.dataset.row = seat.row;
    seatDiv.dataset.number = seat.number;
    seatDiv.dataset.type = seat.type;
    seatDiv.dataset.status = seat.status;
    seatDiv.textContent = seat.number;
    
    // Add click handler for available seats
    if (seat.status === 'available') {
        seatDiv.addEventListener('click', () => toggleSeatSelection(seat));
    }
    
    return seatDiv;
}

    // Toggle seat selection
    function toggleSeatSelection(seat) {
        const index = state.selectedSeats.findIndex(s => s.id === seat.id);
        
        if (index === -1) {
            // Add seat to selection
            state.selectedSeats.push(seat);
            document.querySelector(`.seat[data-id="${seat.id}"]`).classList.add('selected');
        } else {
            // Remove seat from selection
            state.selectedSeats.splice(index, 1);
            document.querySelector(`.seat[data-id="${seat.id}"]`).classList.remove('selected');
        }
        
        updateSelectedSeatsDisplay();
    }

    // Update the display of selected seats and total
    function updateSelectedSeatsDisplay() {
        selectedSeatsList.innerHTML = '';
        
        state.totalAmount = 0;
        
        state.selectedSeats.forEach(seat => {
            const price = PRICES[seat.type];
            state.totalAmount += price;
            
            const li = document.createElement('li');
            li.innerHTML = `
                ${seat.row}${seat.number} (${seat.type === 'vip' ? 'VIP' : 'Regular'}) - ${formatPrice(price)}
                <button class="remove-seat" data-id="${seat.id}">×</button>
            `;
            selectedSeatsList.appendChild(li);
        });
        
        totalAmountElement.textContent = formatPrice(state.totalAmount);
        
        // Add event listeners to remove buttons
        document.querySelectorAll('.remove-seat').forEach(button => {
            button.addEventListener('click', (e) => {
                const seatId = parseInt(e.target.dataset.id);
                const seat = state.seats.find(s => s.id === seatId);
                toggleSeatSelection(seat);
            });
        });
    }

    // Handle manual seat entry
    function handleManualSeatEntry() {
        const input = seatInput.value.trim();
        if (!input) return;
        
        // Parse input (format: A1, B5, etc.)
        const parts = input.split(',').map(part => part.trim());
        
        parts.forEach(part => {
            // Extract row letter and seat number
            const match = part.match(/([A-L])(\d+)/i);
            if (match) {
                const [, row, numStr] = match;
                const number = parseInt(numStr);
                
                // Find the seat in our state
                const seat = state.seats.find(s => 
                    s.row.toUpperCase() === row.toUpperCase() && 
                    s.number === number &&
                    s.status === 'available'
                );
                
                if (seat) {
                    // Only add if not already selected
                    if (!state.selectedSeats.some(s => s.id === seat.id)) {
                        toggleSeatSelection(seat);
                    }
                } else {
                    alert(`Seat ${row}${number} is not available or does not exist.`);
                }
            } else {
                alert(`Invalid seat format: ${part}. Please use format like A1, B5, etc.`);
            }
        });
        
        // Clear input
        seatInput.value = '';
    }

    // Handle form submission
    function handleFormSubmit(e) {
        e.preventDefault();
        
        if (state.selectedSeats.length === 0) {
            alert('Please select at least one seat.');
            return;
        }
        
        const customerName = document.getElementById('customer-name').value;
        const customerEmail = document.getElementById('customer-email').value;
        const customerPhone = document.getElementById('customer-phone').value;
        
        if (!customerName || !customerEmail || !customerPhone) {
            alert('Please fill in all customer information.');
            return;
        }
        
        // Get selected payment method
        let paymentMethod;
        for (const radio of paymentMethodRadios) {
            if (radio.checked) {
                paymentMethod = radio.value;
                break;
            }
        }
        
        // Create booking object
        const bookingData = {
            customerName,
            customerEmail,
            customerPhone,
            paymentMethod,
            seatIds: state.selectedSeats.map(seat => seat.id),
            totalAmount: state.totalAmount
        };
        
        // Send booking to server
        fetch('/api/book', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(bookingData),
        })
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                showBookingConfirmation(data.data.bookingId);
            } else {
                alert(`Booking failed: ${data.message}`);
            }
        })
        .catch(error => {
            console.error('Error:', error);
            alert('Booking failed. Please try again.');
        });
    }

    // Show booking confirmation
    function showBookingConfirmation(bookingId) {
        // Generate a booking reference based on the booking ID
        const bookingRef = 'BK' + bookingId.toString().padStart(4, '0');
        
        // Update UI to show confirmation
        document.getElementById('booking-reference').textContent = bookingRef;
        
        const seatsText = state.selectedSeats.map(seat => 
            `${seat.row}${seat.number} (${seat.type === 'vip' ? 'VIP' : 'Regular'})`
        ).join(', ');
        document.getElementById('confirmation-seats').textContent = seatsText;
        
        document.getElementById('confirmation-amount').textContent = formatPrice(state.totalAmount);
        
        // Show payment details based on method
        const paymentDetails = document.getElementById('confirmation-payment-details');
        const paymentMethod = document.querySelector('input[name="payment-method"]:checked').value;
        
        if (paymentMethod === 'bank_transfer') {
            paymentDetails.innerHTML = `
                <h4>Bank Transfer Details:</h4>
                <p>Bank: BCA</p>
                <p>Account Number: 2920946715</p>
                <p>Account Name: Arif Saputra</p>
                <b>Please include your booking reference (${bookingRef}) in the transfer description.</b>
            `;
        } else {
            paymentDetails.innerHTML = `
                <h4>Cash Payment Details:</h4>
                <p>Visit our booth after school hours to complete your payment.</p>
                <p>Contact: Kezia Orlie (+62 821-2107-8944)</p>
                <p>Contact: Stefanie Christensia Siwu (+62 895-4176-40808)</p>
                <p>Please mention your booking reference (${bookingRef}) when making the payment.</p>
            `;
        }
        
        // Hide booking form and show confirmation
        bookingForm.style.display = 'none';
        bookingConfirmation.style.display = 'block';
        
        // Clear selected seats
        state.selectedSeats = [];
        
        // Re-fetch seats to show updated status
        fetchSeats();
    }

    // Format price as IDR
    function formatPrice(price) {
        return new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR' }).format(price);
    }

    // Payment method toggle
    function togglePaymentInfo() {
        if (paymentMethodRadios[0].checked) {
            bankTransferInfo.style.display = 'block';
            cashInfo.style.display = 'none';
        } else {
            bankTransferInfo.style.display = 'none';
            cashInfo.style.display = 'block';
        }
    }

    // Reset the booking form
    function resetBookingForm() {
        customerForm.reset();
        bookingForm.style.display = 'block';
        bookingConfirmation.style.display = 'none';
        
        // Reset selected seats
        state.selectedSeats = [];
        updateSelectedSeatsDisplay();
        
        // Scroll to top
        window.scrollTo(0, 0);
    }

    // Initialize the app
    function init() {
        fetchSeats();
        
        // Event listeners
        enterSeatsButton.addEventListener('click', handleManualSeatEntry);
        seatInput.addEventListener('keypress', (e) => {
            if (e.key === 'Enter') {
                e.preventDefault();
                handleManualSeatEntry();
            }
        });
        
        customerForm.addEventListener('submit', handleFormSubmit);
        
        paymentMethodRadios.forEach(radio => {
            radio.addEventListener('change', togglePaymentInfo);
        });
        
        newBookingButton.addEventListener('click', resetBookingForm);
    }

    // Start the application
    init();
});