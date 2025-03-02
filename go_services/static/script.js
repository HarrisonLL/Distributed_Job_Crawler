// ------------------------ left panel ------------------------
function setDefaultDate() {
    const today = new Date().toISOString().split('T')[0];
    document.getElementById("startDate").value = today;
    document.getElementById("endDate").value = today;
}

function fetchJobStatsByDate() {
    const startDate = document.getElementById("startDate").value;
    const endDate = document.getElementById("endDate").value;
    fetchJobStats(startDate, endDate);
}

async function fetchJobStats(startDate, endDate) {
    try {
        const response = await fetch(`/api/v1/task_stats?start_date=${startDate}&end_date=${endDate}`);
        const data = await response.json();
        if (!data || data.length === 0) {
            console.warn("No job stats available. Displaying empty state.");
            return;
        }
        renderChart(data);
    } catch (error) {
        console.error("Error fetching job stats:", error);
    }
}

let chartInstance = null; 
function renderChart(jobStats) {
    if (chartInstance) chartInstance.destroy();

    const ctx = document.getElementById('jobChart').getContext('2d');
    const colorMap = generateColorMap(jobStats);

    const labels = [...new Set(jobStats.map(stat => stat.time_period))].sort();
    const datasets = jobStats.map(stat => ({
        label: `${stat.company} - ${stat.job_type}`,
        backgroundColor: colorMap[stat.company].shades[stat.job_type],
        data: labels.map(label => (stat.time_period === label ? stat.job_count : 0))
    }));

    chartInstance = new Chart(ctx, {
        type: 'bar',
        data: { labels, datasets },
        options: {
            responsive: true,
            plugins: { legend: { display: false } },
            scales: { y: { beginAtZero: true } }
        }
    });

    renderCustomLegend(colorMap);
}

function generateColorMap(jobStats) {
    const companies = [...new Set(jobStats.map(stat => stat.company))];
    const jobTypes = [...new Set(jobStats.map(stat => stat.job_type))];
    const baseColors = ["#FF8C00", "#4682B4", "#FF6F61", "#6B8E23", "#FFD700", "#20B2AA", "#DC143C", "#8A2BE2", "#2E8B57", "#1E90FF", "#9932CC", "#FF4500"];

    const colorMap = {};
    companies.forEach((company, index) => {
        colorMap[company] = { baseColor: baseColors[index % baseColors.length], shades: {} };
        jobTypes.forEach((job, jIndex) => {
            colorMap[company].shades[job] = adjustColor(colorMap[company].baseColor, jIndex * 30);
        });
    });

    return colorMap;
}

function adjustColor(baseColor, offset) {
    let [r, g, b] = baseColor.match(/\w\w/g).map(hex => parseInt(hex, 16));
    return `rgb(${Math.min(255, r + offset)},${Math.min(255, g + offset)},${Math.min(255, b + offset)})`;
}

function renderCustomLegend(colorMap) {
    const legendContainer = document.getElementById('customLegend');
    legendContainer.innerHTML = '';

    Object.keys(colorMap).forEach(company => {
        const row = document.createElement('div');
        row.classList.add('legend-row');

        const companyLabel = document.createElement('span');
        companyLabel.innerHTML = `<strong>${company}:</strong> `;
        row.appendChild(companyLabel);

        Object.entries(colorMap[company].shades).forEach(([job, color]) => {
            const colorBox = document.createElement('span');
            colorBox.classList.add('legend-color-box');
            colorBox.style.backgroundColor = color;
            row.appendChild(colorBox);

            const text = document.createElement('span');
            text.innerText = ` ${job} `;
            row.appendChild(text);
        });

        legendContainer.appendChild(row);
    });
}

// ------------------------ right panel ------------------------
async function handleRegister(event) {
    event.preventDefault();
    const username = document.getElementById("registerUsername").value;
    const password = document.getElementById("registerPassword").value;
    const confirmPassword = document.getElementById("confirmPassword").value;

    if (password !== confirmPassword) return alert("Passwords do not match!");

    try {
        const response = await fetch("/api/v1/register", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ username, password })
        });

        const data = await response.json();
        alert(data.message || "Registration failed!");
    } catch (error) {
        console.error("Error during registration:", error);
    }
}

async function handleLogin(event) {
    event.preventDefault();
    const username = document.getElementById("loginUsername").value;
    const password = document.getElementById("loginPassword").value;

    try {
        const response = await fetch("/api/v1/login", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ username, password })
        });

        const data = await response.json();
        if (response.ok) {
            localStorage.setItem("token", data.token);
            alert("Login successful!");
            showPreferencesForm();
            fetchUserProfile();
        } else {
            alert(data.error || "Login failed.");
        }
    } catch (error) {
        console.error("Error during login:", error);
    }
}

document.getElementById("logoutBtn").addEventListener("click", function () {
    localStorage.removeItem("token");
    alert("Logged out successfully!");
    window.location.reload();
});

function showPreferencesForm() {
    document.getElementById("preferencesSection").style.display = "block";
    document.querySelector(".preferences").style.display = "none";
}

async function fetchUserProfile() {
    const token = localStorage.getItem("token");
    if (!token) return;

    try {
        const response = await fetch("/api/v1/user_profile", {
            method: "GET",
            headers: { "Authorization": `Bearer ${token}` }
        });

        const data = await response.json();
        if (response.ok) {
            document.getElementById("yoe").value = data.yoe || "";
            document.getElementById("jobType").value = data.job_type || "";
            document.getElementById("company").value = data.company || "";
        }
    } catch (error) {
        console.error("Error fetching user profile:", error);
    }
}

async function fetchUserProfile() {
    const token = localStorage.getItem("token");
    if (!token) return;
    try {
        const response = await fetch("/api/v1/user_profile", {
            method: "GET",
            headers: {
                "Authorization": `Bearer ${token}`
            }
        });
        const data = await response.json();
        if (response.ok) {
            document.getElementById("email").value = data.email || "";
            document.getElementById("yoe").value = data.yoe || "0-1";
            const selectedCompanies = data.company ? data.company.split(", ") : [];
            document.querySelectorAll("#companyCheckboxes input").forEach(checkbox => {
                checkbox.checked = selectedCompanies.includes(checkbox.value.toLowerCase());
            });
            const selectedJobTypes = data.job_type ? data.job_type.split(", ") : [];
            document.querySelectorAll("#jobTypeCheckboxes input").forEach(checkbox => {
                checkbox.checked = selectedJobTypes.includes(checkbox.value.toLowerCase());
            });
            document.getElementById("emailSubscription").checked = data.email_subscription;
        } else {
            console.error("Failed to fetch user profile:", data.error);
        }
    } catch (error) {
        console.error("Error fetching user profile:", error);
    }
}

document.getElementById("updatePreferencesForm").addEventListener("submit", async function(event) {
    event.preventDefault();
    const token = localStorage.getItem("token");
    if (!token) {
        alert("Please log in first.");
        return;
    }
    const email = document.getElementById("email").value;
    const yoe = document.getElementById("yoe").value;
    const selectedCompanies = Array.from(document.querySelectorAll("#companyCheckboxes input:checked"))
        .map(checkbox => checkbox.value.toLowerCase())
        .join(", ");
    const selectedJobTypes = Array.from(document.querySelectorAll("#jobTypeCheckboxes input:checked"))
        .map(checkbox => checkbox.value.toLowerCase())
        .join(", ");
    const isSubscribed = document.getElementById("emailSubscription").checked;
    try {
        const response = await fetch("/api/v1/update_profile", {
            method: "PATCH",
            headers: {
                "Content-Type": "application/json",
                "Authorization": `Bearer ${token}`
            },
            body: JSON.stringify({ email, yoe, company: selectedCompanies, job_type: selectedJobTypes, email_subscription: isSubscribed })
        });
        const data = await response.json();
        if (response.ok) {
            alert("Profile updated successfully!");
        } else {
            alert(data.error || "Failed to update profile.");
        }
    } catch (error) {
        console.error("Error updating profile:", error);
        alert("An error occurred. Please try again.");
    }
});

function showForm(formId) {
    document.getElementById('loginForm').classList.add('hidden');
    document.getElementById('registerForm').classList.add('hidden');
    document.getElementById(formId).classList.remove('hidden');
    document.querySelectorAll('.tab-btn').forEach(btn => btn.classList.remove('active'));
    document.querySelector(`[data-form="${formId}"]`).classList.add('active');
}

function handleAuthState() {
    const token = localStorage.getItem("token");
    if (token) {
        showPreferencesForm();
        fetchUserProfile();
        document.getElementById("updatePreferencesForm")?.addEventListener("submit", updateUserPreferences);
    }
}

function initializeUI() {
    document.getElementById("fetchDataBtn").addEventListener("click", fetchJobStatsByDate);
    document.getElementById("registerForm").addEventListener("submit", handleRegister);
    document.getElementById("loginForm").addEventListener("submit", handleLogin);
    setDefaultDate();

    document.querySelectorAll(".tab-btn").forEach(button => {
        button.addEventListener("click", function() {
            showForm(this.getAttribute("data-form"));
        });
    });
}

document.addEventListener("DOMContentLoaded", () => {
    console.log("Page loaded, initializing UI...");
    handleAuthState();
    initializeUI();
    fetchJobStatsByDate();
});
