document.addEventListener("DOMContentLoaded", () => {
    Vue.createApp({
      data() {
        return {
          startDate: new Date().toISOString().split("T")[0],
          endDate: new Date().toISOString().split("T")[0],
          jobStats: [],
          isLoggedIn: false,
          token: localStorage.getItem("token") || "",
          activeTab: "login",
          loginData: { username: "", password: "" },
          registerData: { username: "", password: "", confirmPassword: "" },
          userProfile: {
            email: "",
            yoe: "0-1",
            companies: [],
            jobTypes: [],
            emailSubscription: false
          },
          companyList: ["amazon", "meta", "google", "uber"],
          jobTypeList: ["software engineer", "data scientist", "machine learning engineer"],
          chart: null,
          baseColors: ["#FF8C00", "#4682B4", "#FF6F61", "#6B8E23", "#FFD700", "#20B2AA", "#DC143C", "#8A2BE2", "#2E8B57", "#1E90FF", "#9932CC", "#FF4500"]
        };
      },
      mounted() {
        this.fetchJobStats();
        if (this.token) {
          this.isLoggedIn = true;
          this.fetchUserProfile();
        }
      },
      methods: {

        async fetchJobStats() {
          try {
            const response = await fetch(`/api/v1/task_stats?start_date=${this.startDate}&end_date=${this.endDate}`);
            this.jobStats = await response.json();
            this.renderChart();
          } catch (error) {
            console.error("Error fetching job stats:", error);
          }
        },

        adjustColor(baseColor, offset) {
            let [r, g, b] = baseColor.match(/\w\w/g).map(hex => parseInt(hex, 16));
            return `rgb(${Math.min(255, r + offset)},${Math.min(255, g + offset)},${Math.min(255, b + offset)})`;
          },

        renderCustomLegend(colorMap) {
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
        },

        renderChart() {
          if (this.chart) this.chart.destroy();
          const ctx = document.getElementById("jobChart").getContext("2d");
          const companies = [...new Set(this.jobStats.map(stat => stat.company))];
          const jobTypes = [...new Set(this.jobStats.map(stat => stat.job_type))];
          let colorMap = {};
          companies.forEach((company, index) => {
            colorMap[company] = { baseColor: this.baseColors[index % this.baseColors.length], shades: {} };
            jobTypes.forEach((job, jIndex) => {
                colorMap[company].shades[job] = this.adjustColor(colorMap[company].baseColor, jIndex * 30);
            });
          });
          const labels = [...new Set(this.jobStats.map(stat => stat.time_period))].sort();
          const datasets = this.jobStats.map(stat => ({
            label: `${stat.company} - ${stat.job_type}`,
            backgroundColor: colorMap[stat.company].shades[stat.job_type],
            data: labels.map(label => stat.time_period === label ? stat.job_count : 0),
        }));
          this.chart = new Chart(ctx, {
            type: "bar",
            data: { labels, datasets },
            options: {
                responsive: true,
                plugins: { legend: { display: false } },
                scales: { y: { beginAtZero: true } }
            }
          });
          this.renderCustomLegend(colorMap);
        },

        async handleLogin() {
          try {
            const response = await fetch("/api/v1/login", {
              method: "POST",
              headers: { "Content-Type": "application/json" },
              body: JSON.stringify(this.loginData),
            });
  
            const data = await response.json();
            if (response.ok) {
              this.token = data.token;
              localStorage.setItem("token", this.token);
              this.isLoggedIn = true;
              this.fetchUserProfile();
            } else {
              alert(data.message || "Login failed.");
            }
          } catch (error) {
            console.error("Error during login:", error);
          }
        },

        async handleRegister() {
            if (this.registerData.password !== this.registerData.confirmPassword) {
                alert("Passwords do not match!");
                return;
            }
            const response = await fetch("/api/v1/register", {
              method: "POST",
              headers: { "Content-Type": "application/json" },
              body: JSON.stringify(this.registerData),
            });
            const data = await response.json();
            alert(data.message || "Registration failed.");
          
        },

        async fetchUserProfile() {
            if (!this.token) return;
            const response = await fetch("/api/v1/user_profile", {
              method: "GET",
              headers: { "Authorization": `Bearer ${this.token}` },
            });
            const data = await response.json();
            if (response.ok) {
              this.userProfile.email = data.email || "";
              this.userProfile.yoe = data.yoe || "0-1";
              this.userProfile.companies = data.company ? data.company.split(", ") : [];
              this.userProfile.jobTypes = data.job_type ? data.job_type.split(", ") : [];
              this.userProfile.emailSubscription = data.email_subscription || false;
            } else {
              console.error("Failed to fetch user profile:", data.error);
            }
        },

        async updatePreferences() {
            if (!this.token) {
                alert("Please log in first.");
                return;
            }
            const response = await fetch("/api/v1/update_profile", {
              method: "PATCH",
              headers: {
                "Content-Type": "application/json",
                "Authorization": `Bearer ${this.token}`,
              },
              body: JSON.stringify({
                email: this.userProfile.email,
                yoe: this.userProfile.yoe,
                company: this.userProfile.companies.join(", "),
                job_type: this.userProfile.jobTypes.join(", "),
                email_subscription: this.userProfile.emailSubscription,
              }),
            });
  
            const data = await response.json();
            if (response.ok) {
              alert("Profile updated successfully!");
            } else {
              alert(data.error || "Failed to update profile.");
            }
        },

        handleLogout() {
          localStorage.removeItem("token");
          this.token = "";
          this.isLoggedIn = false;
          alert("Logged out successfully!");
        }
      }
    }).mount("#app");
  });
