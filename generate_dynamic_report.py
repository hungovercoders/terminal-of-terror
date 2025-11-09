#!/usr/bin/env python3
# filepath: generate_dynamic_report.py

import csv
import json
import os
import subprocess
from datetime import datetime
import re

def run_lizard_analysis():
    """Run Lizard and parse the output"""
    print("Running Lizard analysis...")
    
    # Run lizard with CSV output
    result = subprocess.run(['lizard', '-l', 'go', '--csv', '.'], 
                          capture_output=True, text=True)
    
    print(f"Lizard return code: {result.returncode}")
    
    if result.returncode != 0:
        print(f"Lizard error: {result.stderr}")
        return run_lizard_text_output()
    
    # Parse CSV output from stdout
    data = []
    lines = result.stdout.strip().split('\n')
    
    if not lines or not lines[0]:
        print("No CSV data found, trying text output...")
        return run_lizard_text_output()
    
    # Lizard CSV format: NLOC,CCN,token,PARAM,length,location,filename,function_name,function_signature,start_line,end_line
    # We need to manually create the header since Lizard doesn't include it
    csv_header = "nloc,ccn,tokens,params,length,location,filename,function_name,function_signature,start_line,end_line"
    csv_data = csv_header + '\n' + result.stdout.strip()
    
    try:
        reader = csv.DictReader(csv_data.split('\n'))
        for row in reader:
            # Clean up function name - remove parameters if present
            func_name = row.get('function_name', '').strip()
            if not func_name or func_name == '':
                # For anonymous functions, try to extract from function_signature
                func_sig = row.get('function_signature', '')
                if func_sig and '(' in func_sig:
                    func_name = func_sig.split('(')[0].strip()
                else:
                    func_name = '<anonymous>'
            
            # Clean up file path
            file_path = row.get('filename', '').strip()
            
            data.append({
                'file': file_path,
                'function': func_name,
                'complexity': int(row.get('ccn', 0)),
                'loc': int(row.get('nloc', 0)),
                'tokens': int(row.get('tokens', 0)),
                'params': int(row.get('params', 0))
            })
            
    except Exception as e:
        print(f"CSV parsing error: {e}")
        print("Falling back to text output parsing...")
        return run_lizard_text_output()
    
    print(f"Successfully parsed {len(data)} functions from CSV")
    return data

def run_lizard_text_output():
    """Alternative approach using lizard text output"""
    print("Using lizard text output parsing...")
    
    # Run lizard with standard output
    result = subprocess.run(['lizard', '-l', 'go', '.'], 
                          capture_output=True, text=True)
    
    if result.returncode != 0:
        print(f"Lizard error: {result.stderr}")
        return []
    
    output = result.stdout
    print(f"Lizard text output (first 500 chars):\n{output[:500]}")
    
    data = []
    lines = output.split('\n')
    
    # Parse the text output
    # Looking for lines like:    6      1     46      0       6 @14-19@./cmd/random.go
    for line in lines:
        line = line.strip()
        
        # Skip empty lines, headers, and summary lines
        if not line or line.startswith('=') or line.startswith('-') or 'NLOC' in line or 'file analyzed' in line:
            continue
            
        # Look for function data lines
        # Format: NLOC CCN token PARAM length location
        # Example:    6      1     46      0       6 @14-19@./cmd/random.go
        #         3      1     12      0       3 init@22-24@./cmd/random.go
        
        # Use regex to parse the line
        match = re.match(r'^\s*(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(.+)$', line)
        if match:
            nloc, ccn, tokens, params, length, location = match.groups()
            
            # Parse location to get function name and file
            # Format can be: @14-19@./cmd/random.go or init@22-24@./cmd/random.go
            if '@' in location:
                parts = location.split('@')
                if len(parts) >= 3:
                    func_name = parts[0] if parts[0] else '<anonymous>'
                    file_path = parts[2]
                else:
                    func_name = '<anonymous>'
                    file_path = location
            else:
                func_name = '<anonymous>'
                file_path = location
            
            data.append({
                'file': file_path,
                'function': func_name,
                'complexity': int(ccn),
                'loc': int(nloc),
                'tokens': int(tokens),
                'params': int(params)
            })
    
    print(f"Text parsing found {len(data)} functions")
    return data

def generate_html_report(data):
    """Generate enhanced HTML report"""
    
    if not data:
        print("No data to generate report from!")
        return create_empty_report()
    
    # Calculate statistics
    total_files = len(set(item['file'] for item in data if item['file']))
    total_functions = len(data)
    avg_complexity = sum(item['complexity'] for item in data) / len(data) if data else 0
    total_loc = sum(item['loc'] for item in data)
    max_complexity = max(item['complexity'] for item in data) if data else 0
    
    # Complexity distribution
    complexity_ranges = {'low': 0, 'medium': 0, 'high': 0}
    for item in data:
        if item['complexity'] <= 3:
            complexity_ranges['low'] += 1
        elif item['complexity'] <= 7:
            complexity_ranges['medium'] += 1
        else:
            complexity_ranges['high'] += 1
    
    # Generate insights
    insights = []
    if max_complexity <= 5:
        insights.append("✅ Excellent: All functions have low complexity (≤5)")
    else:
        complex_funcs = [item for item in data if item['complexity'] > 5]
        insights.append(f"⚠️ {len(complex_funcs)} function(s) have higher complexity")
        
        # Show top complex functions
        top_complex = sorted(complex_funcs, key=lambda x: x['complexity'], reverse=True)[:3]
        for func in top_complex:
            insights.append(f"🔴 {func['function']} (complexity: {func['complexity']})")
    
    avg_loc = total_loc / total_functions if total_functions else 0
    insights.append(f"📏 Average function length: {avg_loc:.1f} lines")
    insights.append("🎯 Target: Keep complexity ≤ 5 for maintainability")
    
    # Add Go-specific insights
    if any('main' in item['function'] for item in data):
        insights.append("🏠 Main functions detected - consider breaking down if complex")
    
    init_funcs = [item for item in data if 'init' in item['function'].lower()]
    if init_funcs:
        insights.append(f"🔧 {len(init_funcs)} init function(s) found")
    
    # Create HTML with proper table generation
    table_rows = generate_table_rows(data)
    
    html_content = f"""<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Terminal of Terror - Code Complexity Analysis</title>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/Chart.js/3.9.1/chart.min.js"></script>
    <style>
        * {{ margin: 0; padding: 0; box-sizing: border-box; }}
        
        body {{
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            background: linear-gradient(135deg, #1a1a2e 0%, #16213e 50%, #0f3460 100%);
            min-height: 100vh;
            color: #333;
        }}
        
        .container {{ max-width: 1200px; margin: 0 auto; padding: 20px; }}
        
        .header {{
            background: rgba(255, 255, 255, 0.95);
            border-radius: 15px;
            padding: 30px;
            margin-bottom: 30px;
            box-shadow: 0 10px 30px rgba(0, 0, 0, 0.3);
            text-align: center;
        }}
        
        .header h1 {{
            color: #2c3e50;
            font-size: 2.5em;
            margin-bottom: 10px;
            text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.1);
        }}
        
        .stats-grid {{
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
            gap: 20px;
            margin-bottom: 30px;
        }}
        
        .stat-card {{
            background: rgba(255, 255, 255, 0.95);
            border-radius: 15px;
            padding: 25px;
            text-align: center;
            box-shadow: 0 8px 25px rgba(0, 0, 0, 0.2);
            transition: transform 0.3s ease;
        }}
        
        .stat-card:hover {{ transform: translateY(-5px); }}
        
        .stat-number {{
            font-size: 2.5em;
            font-weight: bold;
            color: #e74c3c;
        }}
        
        .stat-label {{
            color: #7f8c8d;
            font-size: 1.1em;
            margin-top: 10px;
        }}
        
        .chart-container, .complexity-table, .insight-box {{
            background: rgba(255, 255, 255, 0.95);
            border-radius: 15px;
            padding: 25px;
            margin-bottom: 30px;
            box-shadow: 0 8px 25px rgba(0, 0, 0, 0.2);
        }}
        
        .chart-container {{
            height: 400px;
        }}
        
        .chart-container canvas {{
            max-height: 300px;
        }}
        
        .chart-title {{ font-size: 1.5em; color: #2c3e50; margin-bottom: 20px; text-align: center; }}
        
        table {{ width: 100%; border-collapse: collapse; margin-top: 20px; }}
        
        th {{
            background: linear-gradient(135deg, #e74c3c, #c0392b);
            color: white;
            padding: 15px;
            text-align: left;
            font-weight: 600;
        }}
        
        td {{ padding: 12px 15px; border-bottom: 1px solid #eee; }}
        
        tr:nth-child(even) {{ background-color: #f8f9fa; }}
        tr:hover {{ background-color: #e3f2fd; transition: background-color 0.3s ease; }}
        
        .complexity-low {{ background-color: #d4edda !important; color: #155724; }}
        .complexity-medium {{ background-color: #fff3cd !important; color: #856404; }}
        .complexity-high {{ background-color: #f8d7da !important; color: #721c24; }}
        
        .file-path {{ font-family: 'Courier New', monospace; font-size: 0.9em; color: #6c757d; }}
        .function-name {{ font-family: 'Courier New', monospace; font-weight: bold; color: #495057; }}
        
        .insight-item {{
            margin-bottom: 10px;
            padding: 15px;
            background: #f8f9fa;
            border-radius: 8px;
            border-left: 4px solid #e74c3c;
        }}
        
        .footer {{ text-align: center; color: rgba(255, 255, 255, 0.8); margin-top: 40px; }}
        
        .debug-info {{
            background: rgba(255, 255, 255, 0.9);
            border-radius: 10px;
            padding: 15px;
            margin-bottom: 20px;
            font-family: 'Courier New', monospace;
            font-size: 0.9em;
            color: #333;
            max-height: 200px;
            overflow-y: auto;
        }}
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🦇 Terminal of Terror</h1>
            <p>Code Complexity Analysis Report</p>
            <small>Generated on {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}</small>
        </div>

        <div class="stats-grid">
            <div class="stat-card">
                <div class="stat-number">{total_files}</div>
                <div class="stat-label">Files Analyzed</div>
            </div>
            <div class="stat-card">
                <div class="stat-number">{total_functions}</div>
                <div class="stat-label">Functions</div>
            </div>
            <div class="stat-card">
                <div class="stat-number">{avg_complexity:.1f}</div>
                <div class="stat-label">Avg Complexity</div>
            </div>
            <div class="stat-card">
                <div class="stat-number">{total_loc}</div>
                <div class="stat-label">Lines of Code</div>
            </div>
        </div>

        <div class="chart-container">
            <div class="chart-title">Complexity Distribution</div>
            <div id="chartError" style="display: none; color: #e74c3c; text-align: center; margin: 10px;">
                Chart.js failed to load. Showing alternative view below.
            </div>
            <canvas id="complexityChart" width="400" height="300"></canvas>
            <div id="chartFallback" style="display: block;">
                <div style="text-align: center; padding: 20px;">
                    <div style="margin: 15px; padding: 10px; background: #f8f9fa; border-radius: 8px;">
                        <span style="color: #28a745; font-size: 24px;">●</span> 
                        <strong>Low Complexity (1-3):</strong> {complexity_ranges['low']} functions ({(complexity_ranges['low']/total_functions*100):.1f}%)
                    </div>
                    <div style="margin: 15px; padding: 10px; background: #f8f9fa; border-radius: 8px;">
                        <span style="color: #ffc107; font-size: 24px;">●</span> 
                        <strong>Medium Complexity (4-7):</strong> {complexity_ranges['medium']} functions ({(complexity_ranges['medium']/total_functions*100):.1f}%)
                    </div>
                    <div style="margin: 15px; padding: 10px; background: #f8f9fa; border-radius: 8px;">
                        <span style="color: #dc3545; font-size: 24px;">●</span> 
                        <strong>High Complexity (8+):</strong> {complexity_ranges['high']} functions ({(complexity_ranges['high']/total_functions*100):.1f}%)
                    </div>
                    <div style="margin-top: 20px; font-size: 14px; color: #6c757d;">
                        <button onclick="toggleChart()" style="padding: 8px 16px; background: #007bff; color: white; border: none; border-radius: 4px; cursor: pointer;">
                            Try Interactive Chart
                        </button>
                    </div>
                </div>
            </div>
        </div>

        <div class="insight-box">
            <div class="chart-title">📊 Key Insights</div>
            {''.join(f'<div class="insight-item">{insight}</div>' for insight in insights)}
        </div>

        <div class="complexity-table">
            <div class="chart-title">Detailed Function Analysis</div>
            <table>
                <thead>
                    <tr>
                        <th>File</th>
                        <th>Function</th>
                        <th>Complexity</th>
                        <th>LOC</th>
                        <th>Tokens</th>
                        <th>Parameters</th>
                    </tr>
                </thead>
                <tbody>
                    {table_rows}
                </tbody>
            </table>
        </div>

        <div class="footer">
            <p>Generated for Terminal of Terror project</p>
            <p>Complexity analysis helps identify areas for refactoring and optimization</p>
        </div>
    </div>

    <script>
        let chartCreated = false;
        
        function toggleChart() {{
            const canvas = document.getElementById('complexityChart');
            const fallback = document.getElementById('chartFallback');
            const errorDiv = document.getElementById('chartError');
            
            if (chartCreated) {{
                // Toggle between chart and fallback
                if (canvas.style.display === 'none') {{
                    canvas.style.display = 'block';
                    fallback.style.display = 'none';
                }} else {{
                    canvas.style.display = 'none';
                    fallback.style.display = 'block';
                }}
                return;
            }}
            
            // Try to create chart for the first time
            createChart();
        }}
        
        function createChart() {{
            console.log('Attempting to create chart...');
            const complexityRanges = {json.dumps(complexity_ranges)};
            console.log('Complexity ranges:', complexityRanges);
            
            const canvas = document.getElementById('complexityChart');
            const fallback = document.getElementById('chartFallback');
            const errorDiv = document.getElementById('chartError');
            
            if (!canvas) {{
                console.error('Chart canvas not found!');
                return;
            }}
            
            // Check if Chart.js is loaded
            if (typeof Chart === 'undefined') {{
                console.error('Chart.js not loaded!');
                errorDiv.style.display = 'block';
                return;
            }}
            
            try {{
                const ctx = canvas.getContext('2d');
                console.log('Creating chart with data:', complexityRanges);
                
                const chart = new Chart(ctx, {{
                    type: 'doughnut',
                    data: {{
                        labels: ['Low (1-3)', 'Medium (4-7)', 'High (8+)'],
                        datasets: [{{
                            data: [complexityRanges.low, complexityRanges.medium, complexityRanges.high],
                            backgroundColor: ['#28a745', '#ffc107', '#dc3545'],
                            borderWidth: 3,
                            borderColor: '#fff'
                        }}]
                    }},
                    options: {{
                        responsive: true,
                        maintainAspectRatio: true,
                        plugins: {{
                            legend: {{
                                position: 'bottom',
                                labels: {{
                                    padding: 20,
                                    font: {{
                                        size: 14
                                    }}
                                }}
                            }},
                            tooltip: {{
                                callbacks: {{
                                    label: function(context) {{
                                        const label = context.label || '';
                                        const value = context.parsed;
                                        const total = context.dataset.data.reduce((a, b) => a + b, 0);
                                        const percentage = ((value / total) * 100).toFixed(1);
                                        return label + ': ' + value + ' functions (' + percentage + '%)';
                                    }}
                                }}
                            }}
                        }}
                    }}
                }});
                
                console.log('Chart created successfully!');
                chartCreated = true;
                canvas.style.display = 'block';
                fallback.style.display = 'none';
                
            }} catch (error) {{
                console.error('Error creating chart:', error);
                errorDiv.style.display = 'block';
                errorDiv.innerHTML = 'Chart creation failed: ' + error.message;
            }}
        }}
        
        // Try to create chart automatically when page loads
        document.addEventListener('DOMContentLoaded', function() {{
            console.log('DOM loaded');
            // Hide canvas initially
            const canvas = document.getElementById('complexityChart');
            if (canvas) {{
                canvas.style.display = 'none';
            }}
            
            // Auto-try to create chart after a short delay
            setTimeout(createChart, 1000);
        }});
    </script>
</body>
</html>"""
    
    return html_content

def create_empty_report():
    """Create a report when no data is found"""
    return f"""<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Terminal of Terror - Code Complexity Analysis</title>
    <style>
        body {{ font-family: Arial, sans-serif; margin: 40px; text-align: center; }}
        .error {{ color: #e74c3c; font-size: 1.2em; }}
        .suggestion {{ margin-top: 20px; color: #7f8c8d; }}
    </style>
</head>
<body>
    <h1>🦇 Terminal of Terror - Complexity Analysis</h1>
    <div class="error">No complexity data found!</div>
    <div class="suggestion">
        Make sure Lizard is installed and you're running this from the project root directory.
        <br><br>
        Try running: <code>lizard -l go .</code> manually to check if Lizard works.
    </div>
    <p>Generated on {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}</p>
</body>
</html>"""

def generate_table_rows(data):
    """Generate HTML table rows for complexity data"""
    if not data:
        return '<tr><td colspan="6">No data available</td></tr>'
        
    sorted_data = sorted(data, key=lambda x: x['complexity'], reverse=True)
    rows = []
    
    for item in sorted_data:
        complexity_class = 'complexity-low'
        if item['complexity'] > 7:
            complexity_class = 'complexity-high'
        elif item['complexity'] > 3:
            complexity_class = 'complexity-medium'
        
        function_name = item['function'] if item['function'] else '<anonymous>'
        file_path = item['file'] if item['file'] else 'unknown'
        
        rows.append(f"""
            <tr>
                <td class="file-path">{file_path}</td>
                <td class="function-name">{function_name}</td>
                <td class="{complexity_class}">{item['complexity']}</td>
                <td>{item['loc']}</td>
                <td>{item['tokens']}</td>
                <td>{item['params']}</td>
            </tr>
        """)
    
    return ''.join(rows)

def main():
    """Main function to generate the report"""
    print("🦇 Terminal of Terror - Complexity Analysis")
    print("=" * 50)
    
    # Create output directory
    os.makedirs('complexity_analysis', exist_ok=True)
    
    # Run analysis
    data = run_lizard_analysis()
    
    print(f"Found {len(data)} functions to analyze")
    
    # Generate report
    html_content = generate_html_report(data)
    
    # Write to file
    output_file = 'complexity_analysis/terminal_of_terror_complexity.html'
    with open(output_file, 'w') as f:
        f.write(html_content)
    
    print(f"✅ Enhanced complexity report generated: {output_file}")
    print("🌐 Use the following command to open in browser:")
    print(f'    "$BROWSER" "file://{os.path.abspath(output_file)}"')

if __name__ == "__main__":
    main()