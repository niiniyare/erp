package ui

import "github.com/gofiber/fiber/v2"

// ServeForms renders a comprehensive forms demo page
func (h *UIHandler) ServeForms(c *fiber.Ctx) error {
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Forms Demo - Awo ERP</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <link rel="stylesheet" href="/static/css/main.css">
    <style>
        .form-section {
            border: 1px solid #e5e7eb;
            border-radius: 12px;
            padding: 24px;
            margin-bottom: 24px;
            background: white;
            box-shadow: 0 1px 3px 0 rgba(0, 0, 0, 0.1);
            transition: all 0.2s ease-in-out;
        }
        
        .form-section:hover {
            box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
        }
        
        html[data-theme="dark"] .form-section,
        .dark .form-section {
            border-color: #374151;
            background: #1f2937;
            box-shadow: 0 1px 3px 0 rgba(0, 0, 0, 0.3);
        }
        
        .form-demo {
            padding: 20px;
            border: 2px dashed #d1d5db;
            border-radius: 8px;
            margin: 16px 0;
            background: #f9fafb;
            position: relative;
        }
        
        html[data-theme="dark"] .form-demo,
        .dark .form-demo {
            border-color: #4b5563;
            background: #111827;
        }

        /* Theme switching styles */
        html {
            transition: background-color 0.3s ease, color 0.3s ease;
        }
        
        html[data-theme="dark"] {
            background-color: #111827;
            color: #f9fafb;
        }
        
        html[data-theme="dark"] body {
            background-color: #111827;
            color: #f9fafb;
        }

        /* Form validation styles */
        .field-error {
            color: #ef4444;
            font-size: 0.875rem;
            margin-top: 0.25rem;
        }
        
        .field-success {
            color: #10b981;
            font-size: 0.875rem;
            margin-top: 0.25rem;
        }
        
        .input-error {
            border-color: #ef4444 !important;
            box-shadow: 0 0 0 1px #ef4444 !important;
        }
        
        .input-success {
            border-color: #10b981 !important;
            box-shadow: 0 0 0 1px #10b981 !important;
        }
    </style>
</head>
<body class="bg-gray-50 dark:bg-gray-900">
    <div class="min-h-screen flex">
        <!-- Sidebar -->
        <div class="hidden md:flex md:flex-shrink-0">
            <div class="flex flex-col w-64">
                <div class="flex flex-col h-0 flex-1 bg-gray-800">
                    <div class="flex items-center h-16 flex-shrink-0 px-4 bg-gray-900">
                        <span class="text-white text-xl font-semibold">Awo ERP</span>
                    </div>
                    <div class="flex-1 flex flex-col overflow-y-auto">
                        <nav class="flex-1 px-2 py-4 bg-gray-800 space-y-1">
                            <a href="/ui/demo" class="text-gray-300 hover:bg-gray-700 hover:text-white group flex items-center px-2 py-2 text-sm font-medium rounded-md">
                                Dashboard
                            </a>
                            <a href="/ui/demo/components" class="text-gray-300 hover:bg-gray-700 hover:text-white group flex items-center px-2 py-2 text-sm font-medium rounded-md">
                                Components
                            </a>
                            <a href="/ui/demo/forms" class="bg-gray-900 text-white group flex items-center px-2 py-2 text-sm font-medium rounded-md">
                                Forms
                            </a>
                            <a href="/ui/login" class="text-gray-300 hover:bg-gray-700 hover:text-white group flex items-center px-2 py-2 text-sm font-medium rounded-md">
                                Login
                            </a>
                        </nav>
                    </div>
                </div>
            </div>
        </div>

        <!-- Main content -->
        <div class="flex flex-col w-0 flex-1 overflow-hidden">
            <main class="flex-1 relative overflow-y-auto focus:outline-none">
                <div class="py-6">
                    <div class="max-w-7xl mx-auto px-4 sm:px-6 md:px-8">
                        <h1 class="text-3xl font-bold text-gray-900 dark:text-gray-100">Forms Showcase</h1>
                        <p class="mt-2 text-gray-600 dark:text-gray-400">Interactive form examples and validation patterns</p>
                    </div>
                    
                    <div class="max-w-7xl mx-auto px-4 sm:px-6 md:px-8 mt-8">
                        
                        <!-- User Registration Form -->
                        <div class="form-section">
                            <h2 class="text-2xl font-bold text-gray-900 dark:text-gray-100 mb-4">User Registration Form</h2>
                            <div class="form-demo">
                                <form id="registration-form" class="space-y-6">
                                    <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                                        <div class="form-group">
                                            <label for="firstName" class="form-label">First Name *</label>
                                            <input type="text" id="firstName" name="firstName" required 
                                                   class="form-input" placeholder="Enter your first name">
                                        </div>
                                        <div class="form-group">
                                            <label for="lastName" class="form-label">Last Name *</label>
                                            <input type="text" id="lastName" name="lastName" required 
                                                   class="form-input" placeholder="Enter your last name">
                                        </div>
                                    </div>
                                    
                                    <div class="form-group">
                                        <label for="email" class="form-label">Email Address *</label>
                                        <input type="email" id="email" name="email" required 
                                               class="form-input" placeholder="Enter your email address">
                                    </div>
                                    
                                    <div class="form-group">
                                        <label for="password" class="form-label">Password *</label>
                                        <input type="password" id="password" name="password" required 
                                               minlength="8" class="form-input" placeholder="Enter a secure password">
                                        <p class="text-sm text-gray-500 dark:text-gray-400 mt-1">Must be at least 8 characters long</p>
                                    </div>
                                    
                                    <div class="form-group">
                                        <label for="confirmPassword" class="form-label">Confirm Password *</label>
                                        <input type="password" id="confirmPassword" name="confirmPassword" required 
                                               class="form-input" placeholder="Confirm your password">
                                    </div>
                                    
                                    <div class="form-group">
                                        <label for="role" class="form-label">Role</label>
                                        <select id="role" name="role" class="form-input">
                                            <option value="">Select a role</option>
                                            <option value="admin">Administrator</option>
                                            <option value="manager">Manager</option>
                                            <option value="user">User</option>
                                            <option value="viewer">Viewer</option>
                                        </select>
                                    </div>
                                    
                                    <div class="form-group">
                                        <label for="bio" class="form-label">Bio</label>
                                        <textarea id="bio" name="bio" rows="4" class="form-input" 
                                                  placeholder="Tell us about yourself..."></textarea>
                                    </div>
                                    
                                    <div class="form-group">
                                        <div class="flex items-center space-x-2">
                                            <input type="checkbox" id="terms" name="terms" required 
                                                   class="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded">
                                            <label for="terms" class="text-sm text-gray-700 dark:text-gray-300">
                                                I agree to the <a href="#" class="text-blue-600 hover:text-blue-500">Terms and Conditions</a> *
                                            </label>
                                        </div>
                                    </div>
                                    
                                    <div class="form-group">
                                        <div class="flex items-center space-x-2">
                                            <input type="checkbox" id="newsletter" name="newsletter" 
                                                   class="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded">
                                            <label for="newsletter" class="text-sm text-gray-700 dark:text-gray-300">
                                                Subscribe to our newsletter
                                            </label>
                                        </div>
                                    </div>
                                    
                                    <div class="flex justify-end space-x-4">
                                        <button type="reset" class="px-4 py-2 border border-gray-300 rounded-md text-gray-700 bg-white hover:bg-gray-50 dark:bg-gray-800 dark:border-gray-600 dark:text-gray-300 dark:hover:bg-gray-700">
                                            Reset
                                        </button>
                                        <button type="submit" class="px-6 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2">
                                            Register
                                        </button>
                                    </div>
                                </form>
                            </div>
                        </div>

                        <!-- Contact Form -->
                        <div class="form-section">
                            <h2 class="text-2xl font-bold text-gray-900 dark:text-gray-100 mb-4">Contact Form</h2>
                            <div class="form-demo">
                                <form id="contact-form" class="space-y-6">
                                    <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                                        <div class="form-group">
                                            <label for="contactName" class="form-label">Full Name *</label>
                                            <input type="text" id="contactName" name="contactName" required 
                                                   class="form-input" placeholder="Your full name">
                                        </div>
                                        <div class="form-group">
                                            <label for="contactEmail" class="form-label">Email *</label>
                                            <input type="email" id="contactEmail" name="contactEmail" required 
                                                   class="form-input" placeholder="your.email@example.com">
                                        </div>
                                    </div>
                                    
                                    <div class="form-group">
                                        <label for="subject" class="form-label">Subject *</label>
                                        <input type="text" id="subject" name="subject" required 
                                               class="form-input" placeholder="What is this about?">
                                    </div>
                                    
                                    <div class="form-group">
                                        <label for="priority" class="form-label">Priority</label>
                                        <div class="space-y-2">
                                            <div class="flex items-center space-x-2">
                                                <input type="radio" id="priority-low" name="priority" value="low" 
                                                       class="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300">
                                                <label for="priority-low" class="text-sm text-gray-700 dark:text-gray-300">Low</label>
                                            </div>
                                            <div class="flex items-center space-x-2">
                                                <input type="radio" id="priority-medium" name="priority" value="medium" checked
                                                       class="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300">
                                                <label for="priority-medium" class="text-sm text-gray-700 dark:text-gray-300">Medium</label>
                                            </div>
                                            <div class="flex items-center space-x-2">
                                                <input type="radio" id="priority-high" name="priority" value="high" 
                                                       class="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300">
                                                <label for="priority-high" class="text-sm text-gray-700 dark:text-gray-300">High</label>
                                            </div>
                                        </div>
                                    </div>
                                    
                                    <div class="form-group">
                                        <label for="category" class="form-label">Category</label>
                                        <select id="category" name="category" class="form-input">
                                            <option value="">Select a category</option>
                                            <option value="general">General Inquiry</option>
                                            <option value="support">Technical Support</option>
                                            <option value="billing">Billing Question</option>
                                            <option value="feature">Feature Request</option>
                                            <option value="bug">Bug Report</option>
                                        </select>
                                    </div>
                                    
                                    <div class="form-group">
                                        <label for="message" class="form-label">Message *</label>
                                        <textarea id="message" name="message" rows="6" required 
                                                  class="form-input" placeholder="Please describe your inquiry in detail..."></textarea>
                                    </div>
                                    
                                    <div class="flex justify-end space-x-4">
                                        <button type="reset" class="px-4 py-2 border border-gray-300 rounded-md text-gray-700 bg-white hover:bg-gray-50 dark:bg-gray-800 dark:border-gray-600 dark:text-gray-300 dark:hover:bg-gray-700">
                                            Clear
                                        </button>
                                        <button type="submit" class="px-6 py-2 bg-green-600 text-white rounded-md hover:bg-green-700 focus:outline-none focus:ring-2 focus:ring-green-500 focus:ring-offset-2">
                                            Send Message
                                        </button>
                                    </div>
                                </form>
                            </div>
                        </div>

                        <!-- Advanced Form Features -->
                        <div class="form-section">
                            <h2 class="text-2xl font-bold text-gray-900 dark:text-gray-100 mb-4">Advanced Form Features</h2>
                            <div class="form-demo">
                                <form id="advanced-form" class="space-y-6">
                                    <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                                        <div class="form-group">
                                            <label for="dateOfBirth" class="form-label">Date of Birth</label>
                                            <input type="date" id="dateOfBirth" name="dateOfBirth" 
                                                   class="form-input">
                                        </div>
                                        <div class="form-group">
                                            <label for="phoneNumber" class="form-label">Phone Number</label>
                                            <input type="tel" id="phoneNumber" name="phoneNumber" 
                                                   pattern="[0-9]{3}-[0-9]{3}-[0-9]{4}" 
                                                   class="form-input" placeholder="123-456-7890">
                                        </div>
                                    </div>
                                    
                                    <div class="form-group">
                                        <label for="website" class="form-label">Website URL</label>
                                        <input type="url" id="website" name="website" 
                                               class="form-input" placeholder="https://example.com">
                                    </div>
                                    
                                    <div class="form-group">
                                        <label for="experience" class="form-label">Years of Experience</label>
                                        <input type="number" id="experience" name="experience" 
                                               min="0" max="50" class="form-input" placeholder="0">
                                    </div>
                                    
                                    <div class="form-group">
                                        <label for="skills" class="form-label">Skills (Select multiple)</label>
                                        <select id="skills" name="skills" multiple class="form-input" size="5">
                                            <option value="javascript">JavaScript</option>
                                            <option value="python">Python</option>
                                            <option value="go">Go</option>
                                            <option value="react">React</option>
                                            <option value="vue">Vue.js</option>
                                            <option value="node">Node.js</option>
                                            <option value="sql">SQL</option>
                                            <option value="docker">Docker</option>
                                        </select>
                                        <p class="text-sm text-gray-500 dark:text-gray-400 mt-1">Hold Ctrl/Cmd to select multiple options</p>
                                    </div>
                                    
                                    <div class="form-group">
                                        <label for="portfolioFile" class="form-label">Portfolio File</label>
                                        <input type="file" id="portfolioFile" name="portfolioFile" 
                                               accept=".pdf,.doc,.docx" class="form-input">
                                        <p class="text-sm text-gray-500 dark:text-gray-400 mt-1">Accepted formats: PDF, DOC, DOCX</p>
                                    </div>
                                    
                                    <div class="form-group">
                                        <label for="availability" class="form-label">Availability</label>
                                        <div class="space-y-2">
                                            <div class="flex items-center space-x-2">
                                                <input type="checkbox" id="weekdays" name="availability" value="weekdays" 
                                                       class="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded">
                                                <label for="weekdays" class="text-sm text-gray-700 dark:text-gray-300">Weekdays</label>
                                            </div>
                                            <div class="flex items-center space-x-2">
                                                <input type="checkbox" id="weekends" name="availability" value="weekends" 
                                                       class="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded">
                                                <label for="weekends" class="text-sm text-gray-700 dark:text-gray-300">Weekends</label>
                                            </div>
                                            <div class="flex items-center space-x-2">
                                                <input type="checkbox" id="evenings" name="availability" value="evenings" 
                                                       class="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded">
                                                <label for="evenings" class="text-sm text-gray-700 dark:text-gray-300">Evenings</label>
                                            </div>
                                        </div>
                                    </div>
                                    
                                    <div class="flex justify-end space-x-4">
                                        <button type="button" onclick="validateForm()" class="px-4 py-2 border border-blue-600 text-blue-600 rounded-md hover:bg-blue-50 dark:hover:bg-blue-900">
                                            Validate
                                        </button>
                                        <button type="submit" class="px-6 py-2 bg-purple-600 text-white rounded-md hover:bg-purple-700 focus:outline-none focus:ring-2 focus:ring-purple-500 focus:ring-offset-2">
                                            Submit Application
                                        </button>
                                    </div>
                                </form>
                            </div>
                        </div>

                        <!-- Form States Demo -->
                        <div class="form-section">
                            <h2 class="text-2xl font-bold text-gray-900 dark:text-gray-100 mb-4">Form States & Validation</h2>
                            <div class="form-demo">
                                <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                                    <div class="space-y-4">
                                        <h3 class="text-lg font-semibold text-gray-900 dark:text-gray-100">Success State</h3>
                                        <div class="form-group">
                                            <label class="form-label">Valid Email</label>
                                            <input type="email" value="user@example.com" class="form-input input-success" readonly>
                                            <div class="field-success">✓ Email format is valid</div>
                                        </div>
                                        
                                        <div class="form-group">
                                            <label class="form-label">Strong Password</label>
                                            <input type="password" value="SecureP@ssw0rd123" class="form-input input-success" readonly>
                                            <div class="field-success">✓ Password meets all requirements</div>
                                        </div>
                                    </div>
                                    
                                    <div class="space-y-4">
                                        <h3 class="text-lg font-semibold text-gray-900 dark:text-gray-100">Error State</h3>
                                        <div class="form-group">
                                            <label class="form-label">Invalid Email</label>
                                            <input type="email" value="invalid-email" class="form-input input-error" readonly>
                                            <div class="field-error">✗ Please enter a valid email address</div>
                                        </div>
                                        
                                        <div class="form-group">
                                            <label class="form-label">Weak Password</label>
                                            <input type="password" value="123" class="form-input input-error" readonly>
                                            <div class="field-error">✗ Password must be at least 8 characters</div>
                                        </div>
                                    </div>
                                </div>
                                
                                <div class="mt-6 p-4 bg-blue-50 dark:bg-blue-900 rounded-lg">
                                    <h4 class="font-medium text-blue-900 dark:text-blue-100 mb-2">Form Validation Features:</h4>
                                    <ul class="text-sm text-blue-800 dark:text-blue-200 space-y-1">
                                        <li>• Real-time validation on field blur</li>
                                        <li>• Visual feedback with colors and icons</li>
                                        <li>• Custom error messages</li>
                                        <li>• Password strength indication</li>
                                        <li>• Form submission prevention if invalid</li>
                                    </ul>
                                </div>
                            </div>
                        </div>

                    </div>
                </div>
            </main>
        </div>
    </div>
    
    <script src="/static/js/main.js"></script>
    <script>
        // Custom validation for forms demo
        function validateForm() {
            const form = document.getElementById('advanced-form');
            const inputs = form.querySelectorAll('input, select, textarea');
            let validCount = 0;
            let totalCount = 0;
            
            inputs.forEach(input => {
                if (input.type !== 'submit' && input.type !== 'button') {
                    totalCount++;
                    if (input.checkValidity() && input.value.trim() !== '') {
                        validCount++;
                        input.classList.remove('input-error');
                        input.classList.add('input-success');
                    } else {
                        input.classList.remove('input-success');
                        input.classList.add('input-error');
                    }
                }
            });
            
            const percentage = Math.round((validCount / totalCount) * 100);
            window.notifications.show(
                'Form validation: ' + validCount + '/' + totalCount + ' fields valid (' + percentage + '%)', 
                percentage === 100 ? 'success' : 'warning',
                4000
            );
        }
        
        // Custom password confirmation validation
        document.addEventListener('DOMContentLoaded', function() {
            const password = document.getElementById('password');
            const confirmPassword = document.getElementById('confirmPassword');
            
            function validatePasswordMatch() {
                if (confirmPassword.value && password.value !== confirmPassword.value) {
                    confirmPassword.setCustomValidity('Passwords do not match');
                } else {
                    confirmPassword.setCustomValidity('');
                }
            }
            
            if (password && confirmPassword) {
                password.addEventListener('input', validatePasswordMatch);
                confirmPassword.addEventListener('input', validatePasswordMatch);
            }
        });
    </script>
</body>
</html>`

	return c.Type("html").Send([]byte(html))
}
