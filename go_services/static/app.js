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
          userProfile: { email: "", yoe: "0-1", companies: [], jobTypes: [], emailSubscription: false },
          companyList: ["amazon", "google", "meta", "salesforce", "uber"],
          jobTypeList: ["software engineer", "data scientist", "machine learning engineer"],
          selectedJobType: "software engineer",
          chart: null,
          companyColors: { "amazon": "#F79B1B", "google": "#F4B400", "meta": "#4267B2", "salesforce": "#00A1E0", "uber": "#333333"}
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
            const response = await fetch(`/api/v1/task_stats?start_date=${this.startDate}&end_date=${this.endDate}&job_type=${encodeURIComponent(this.selectedJobType)}`);
            this.jobStats = await response.json();
            this.renderChart();
          } catch (error) {
            console.error("Error fetching job stats:", error);
          }
        },
        distroyChart() {
          if (this.chart) {
            this.chart.destroy();
            this.chart = null;
          }
        },
        renderChart() {
          this.distroyChart();
          if (!Array.isArray(this.jobStats) || !this.jobStats.length) return;
          if (!document.getElementById("jobChart")) return;
          const ctx = document.getElementById("jobChart").getContext("2d");
          const companyJobCounts = {};
          this.jobStats.forEach(stat => {
            const company = stat.company;
            const count = stat.job_count || 0;
            companyJobCounts[company] = (companyJobCounts[company] || 0) + count;
          });
          const sortedEntries = Object.entries(companyJobCounts).sort((a, b) => b[1] - a[1]);
          const values = sortedEntries.map(([_, count]) => count);
          const companyColors = this.companyColors;
          const colorSquaresPlugin = {
            id: 'coloredLabels',
            afterDraw(chart) {
              const yAxis = chart.scales.y;
              const ctx = chart.ctx;
              const labels = chart.data.labels;
              labels.forEach((company, index) => {
                const y = yAxis.getPixelForTick(index);
                ctx.fillStyle = companyColors[company];
                ctx.fillRect(10, y - 7, 10, 10);
              });
            }
          };
          this.chart = new Chart(ctx, {
            type: "bar",
            data: {
              labels: sortedEntries.map(([company]) => company),
              datasets: [{ label: "Jobs Found", data: values,}]
            },
            options: {
              indexAxis: 'y',
              responsive: true,
              plugins: {
                legend: { display: false }
              },
              scales: {
                x: { beginAtZero: true,},
                y: {
                  ticks: {
                    font: { size: 14 },
                    callback: function(_, index) {
                      const company = sortedEntries[index][0];
                      return company.charAt(0).toUpperCase() + company.slice(1);
                    }
                  },
                  title: {
                    display: true,
                    padding: {
                      top: 0,
                      bottom: 10
                    }
                  }
                }
              }
            },
            plugins: [colorSquaresPlugin]
          });
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
