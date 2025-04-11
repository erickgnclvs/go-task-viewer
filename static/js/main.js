document.addEventListener('DOMContentLoaded', function() {
    // Tab switching functionality
    const textTab = document.getElementById('text-tab');
    const fileTab = document.getElementById('file-tab');
    const textPanel = document.getElementById('text-panel');
    const filePanel = document.getElementById('file-panel');
    
    textTab.addEventListener('click', function() {
        textTab.classList.add('active');
        fileTab.classList.remove('active');
        textPanel.style.display = 'block';
        filePanel.style.display = 'none';
    });
    
    fileTab.addEventListener('click', function() {
        fileTab.classList.add('active');
        textTab.classList.remove('active');
        filePanel.style.display = 'block';
        textPanel.style.display = 'none';
    });
    
    // File upload functionality
    const dropArea = document.getElementById('drop-area');
    const fileInput = document.getElementById('file-input');
    const fileName = document.getElementById('file-name');
    
    // Prevent defaults for drag events
    ['dragenter', 'dragover', 'dragleave', 'drop'].forEach(eventName => {
        dropArea.addEventListener(eventName, preventDefaults, false);
    });
    
    function preventDefaults(e) {
        e.preventDefault();
        e.stopPropagation();
    }
    
    // Highlight drop area when dragging over it
    ['dragenter', 'dragover'].forEach(eventName => {
        dropArea.addEventListener(eventName, highlight, false);
    });
    
    ['dragleave', 'drop'].forEach(eventName => {
        dropArea.addEventListener(eventName, unhighlight, false);
    });
    
    function highlight() {
        dropArea.classList.add('highlight');
    }
    
    function unhighlight() {
        dropArea.classList.remove('highlight');
    }
    
    // Handle file drop
    dropArea.addEventListener('drop', handleDrop, false);
    
    function handleDrop(e) {
        const dt = e.dataTransfer;
        const files = dt.files;
        fileInput.files = files;
        updateFileName();
    }
    
    // Handle file input change
    fileInput.addEventListener('change', updateFileName);
    
    function updateFileName() {
        if (fileInput.files.length > 0) {
            fileName.textContent = fileInput.files[0].name;
        }
    }
    
    // Add event listener to the toggle details button if it exists
    const toggleButton = document.getElementById('toggleDetails');
    if (toggleButton) {
        toggleButton.addEventListener('click', function() {
            const detailsForm = document.getElementById('detailsForm');
            const showDetailsInput = document.getElementById('showDetailsInput');
            
            // Toggle the value
            if (showDetailsInput.value === 'on') {
                showDetailsInput.value = 'off';
            } else {
                showDetailsInput.value = 'on';
            }
            
            // Submit the form
            detailsForm.submit();
        });
    }
    
    // Modal functionality for How to Use button
    const modal = document.getElementById('howToUseModal');
    const howToUseBtn = document.getElementById('howToUseButton');
    const closeBtn = document.querySelector('.close-button');
    
    // Open modal when button is clicked
    howToUseBtn.addEventListener('click', function() {
        modal.style.display = 'block';
    });
    
    // Close modal when close button is clicked
    closeBtn.addEventListener('click', function() {
        modal.style.display = 'none';
    });
    
    // Close modal when clicking outside of it
    window.addEventListener('click', function(event) {
        if (event.target === modal) {
            modal.style.display = 'none';
        }
    });

    // charts
    const dataContainer = document.getElementById('chartDataContainer');
    if (dataContainer) {
        // Get hour percentages for the hours chart
        const taskHoursPercent = parseFloat(dataContainer.getAttribute('data-task-percent')) || 0;
        const exceededHoursPercent = parseFloat(dataContainer.getAttribute('data-exceeded-percent')) || 0;
        const otherHoursPercent = parseFloat(dataContainer.getAttribute('data-other-percent')) || 0;
        
        // Get values for the values chart
        const taskValue = parseFloat(dataContainer.getAttribute('data-task-value').replace(/[^0-9.]/g, '')) || 0;
        const exceededValue = parseFloat(dataContainer.getAttribute('data-exceeded-value').replace(/[^0-9.]/g, '')) || 0;
        const otherValue = parseFloat(dataContainer.getAttribute('data-other-value').replace(/[^0-9.]/g, '')) || 0;
        
        // Create hours chart
        const hoursChartCtx = document.getElementById('hoursChart').getContext('2d');
        const hoursChart = new Chart(hoursChartCtx, {
            type: 'pie',
            data: {
                labels: ['Tarefas', 'Tempo Excedido', 'Outros'],
                datasets: [{
                    data: [taskHoursPercent, exceededHoursPercent, otherHoursPercent],
                    backgroundColor: [
                        'rgba(67, 97, 238, 0.8)',  // Task (blue)
                        'rgba(237, 137, 54, 0.8)', // Exceeded (orange)
                        'rgba(156, 101, 202, 0.8)'  // Other (purple)
                    ],
                    borderColor: [
                        'rgba(67, 97, 238, 1)',
                        'rgba(237, 137, 54, 1)',
                        'rgba(156, 101, 202, 1)'
                    ],
                    borderWidth: 1
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: {
                        position: 'bottom'
                    },
                    title: {
                        display: true,
                        text: 'Distribuição de Horas por Tipo',
                        font: {
                            size: 16
                        }
                    },
                    tooltip: {
                        callbacks: {
                            label: function(context) {
                                return context.label + ': ' + context.raw.toFixed(2) + '%';
                            }
                        }
                    }
                }
            }
        });
        
        // Create values chart
        const valuesChartCtx = document.getElementById('valuesChart').getContext('2d');
        const valuesChart = new Chart(valuesChartCtx, {
            type: 'pie',
            data: {
                labels: ['Tarefas', 'Tempo Excedido', 'Outros'],
                datasets: [{
                    data: [taskValue, exceededValue, otherValue],
                    backgroundColor: [
                        'rgba(67, 97, 238, 0.8)',  // Task (blue)
                        'rgba(237, 137, 54, 0.8)', // Exceeded (orange)
                        'rgba(156, 101, 202, 0.8)'  // Other (purple)
                    ],
                    borderColor: [
                        'rgba(67, 97, 238, 1)',
                        'rgba(237, 137, 54, 1)',
                        'rgba(156, 101, 202, 1)'
                    ],
                    borderWidth: 1
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: {
                        position: 'bottom'
                    },
                    title: {
                        display: true,
                        text: 'Distribuição de Valores',
                        font: {
                            size: 16
                        }
                    },
                    tooltip: {
                        callbacks: {
                            label: function(context) {
                                return context.label + ': $' + context.raw.toFixed(2);
                            }
                        }
                    }
                }
            }
        });
    }
});