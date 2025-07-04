import subprocess
import sys
import os

sys.path.insert(0, os.path.abspath(os.path.dirname(__file__)))

import signal
import argparse
import time
import psutil
from datetime import datetime
from rich.console import Console
from rich.panel import Panel
from rich.table import Table
from rich.progress import Progress, SpinnerColumn, TextColumn
from rich.prompt import Confirm

try:
    from setup import setup
except ImportError:
    print("setup.py not found. Please ensure it exists in the same directory.")
    sys.exit(1)

console = Console()

class Server:
    def __init__(self, name, module, port, domain, path_prefix=""):
        self.name = name
        self.module = module
        self.port = port
        self.domain = domain
        self.path_prefix = path_prefix
        self.process = None
        self.start_time = None
        self.pid_file = f".{name.lower().replace(' ', '_')}_pid"
    
    def start(self):
        python_path = sys.executable
        cmd = [python_path, "-m", "uvicorn", self.module, "--port", str(self.port), "--log-level", "error", "--host", "127.0.0.1"]
        self.process = subprocess.Popen(cmd)
        self.start_time = datetime.now()
        
        with open(self.pid_file, 'w') as f:
            f.write(str(self.process.pid))
            
        return self.process.pid
    
    def stop(self):
        if self.process:
            self.process.send_signal(signal.SIGINT)
            if os.path.exists(self.pid_file):
                os.remove(self.pid_file)
    
    def check_status(self):
        if os.path.exists(self.pid_file):
            try:
                with open(self.pid_file, 'r') as f:
                    pid = int(f.read().strip())
                
                if psutil.pid_exists(pid):
                    self.process = psutil.Process(pid)
                    if not self.start_time:
                        self.start_time = datetime.fromtimestamp(self.process.create_time())
                    return True
                else:
                    os.remove(self.pid_file)
            except Exception:
                pass
        return False
    
    @property
    def uptime(self):
        if not self.start_time:
            return "N/A"
        delta = datetime.now() - self.start_time
        hours, remainder = divmod(delta.seconds, 3600)
        minutes, seconds = divmod(remainder, 60)
        if delta.days > 0:
            return f"{delta.days}d {hours:02}h:{minutes:02}m"
        return f"{hours:02}:{minutes:02}:{seconds:02}"
    
    @property
    def requests_per_second(self):
        if not self.process or not isinstance(self.process, psutil.Process):
            return "0.00"
        
        try:
            connections = len([c for c in self.process.net_connections() if c.status == 'ESTABLISHED'])
            return f"{connections:.2f}"
        except Exception:
            return "0.00"

def request_sudo_permission():
    if os.geteuid() != 0:
        console.print(Panel.fit(
            "[yellow]⚠️ Bluppi requires sudo privileges to manage system services[/yellow]\n\n"
            "[white]This is needed to start Redis, PostgreSQL, and Nginx services.[/white]",
            title="[bold blue]Sudo Required[/bold blue]", 
            border_style="blue"
        ))
        
        if not Confirm.ask("Continue with sudo?"):
            console.print("[red]Operation cancelled.[/red]")
            sys.exit(1)
            
        os.system('clear')
        
        python_path = sys.executable
        args = ['sudo', python_path] + sys.argv
        os.execvp('sudo', args)

def check_cloudflare_status():
    try:
        output = subprocess.check_output(["pgrep", "-f", "cloudflared.*tunnel.*run"]).decode().strip()
        if output:
            return int(output.split()[0])
    except Exception:
        pass
    return None

def start_cloudflare():
    console.print("[bold blue]Starting Cloudflare tunnel...[/bold blue]")
    
    env = os.environ.copy()
    if 'SUDO_USER' in env:
        original_user = env['SUDO_USER']
        original_home = f"/home/{original_user}"
        env['HOME'] = original_home
        env['USER'] = original_user
    
    return subprocess.Popen(["cloudflared", "tunnel", "run", "bluppi"], env=env)

def get_server_table(servers, cf_pid=None):
    table = Table(show_header=True, header_style="bold magenta")
    table.add_column("Service")
    table.add_column("Status")
    table.add_column("Port")
    table.add_column("Domain/Path")
    table.add_column("PID")
    table.add_column("Uptime")
    table.add_column("Req/sec")
    
    for server in servers:
        is_running = server.check_status()
        status = "[green]Running[/green]" if is_running else "[red]Stopped[/red]"
        pid = str(server.process.pid) if server.process else "N/A"
        domain_path = f"{server.domain}{server.path_prefix}" if server.path_prefix else server.domain
        
        table.add_row(
            server.name,
            status,
            str(server.port),
            domain_path,
            pid,
            server.uptime if is_running else "N/A",
            server.requests_per_second if is_running else "N/A"
        )
    
    if cf_pid:
        table.add_row(
            "Cloudflare Tunnel",
            "[green]Running[/green]",
            "N/A",
            "*.saikat.in",
            str(cf_pid),
            "N/A",
            "N/A"
        )
    
    return table

def main():
    parser = argparse.ArgumentParser(description="Start Bluppi services")
    parser.add_argument("--setup", action="store_true", help="Run setup before starting")
    parser.add_argument("--no-cloudflare", action="store_true", help="Don't start Cloudflare tunnel")
    parser.add_argument("--status", action="store_true", help="Show status without starting servers")
    args = parser.parse_args()
    
    if args.setup:
        setup()
    
    servers = [
        Server("Bluppi API", "app.server:app", 8000, "bluppi-api.saikat.in"),
        Server("Bluppi WS1", "chat.server:app", 8080, "bluppi-ws1.saikat.in"),
        Server("Bluppi WS2", "chat.server:app", 8081, "bluppi-ws2.saikat.in"),
        Server("Bluppi gRPC", "party.server", 50051, "bluppi-grpc.saikat.in"),
    ]
    
    if args.status:
        cf_pid = check_cloudflare_status()
        console.print(get_server_table(servers, cf_pid))
        return
    
    console.print(Panel.fit("[bold blue]Bluppi Server Manager[/bold blue]", 
                          subtitle="Starting services..."))
    
    request_sudo_permission() 
    
    with Progress(
        SpinnerColumn(),
        TextColumn("[progress.description]{task.description}"),
        console=console
    ) as progress:
        task = progress.add_task("[green]Starting services...", total=1)
        
        subprocess.run(["sudo", "systemctl", "start", "redis-server"])
        subprocess.run(["sudo", "systemctl", "start", "postgresql"])
        
        progress.update(task, completed=1)
    
    try:      
        for server in servers:
            console.print(f"[bold green]Starting {server.name} on port {server.port}...[/bold green]")
            server.start()
            time.sleep(1)
        
        cf_pid = None
        if not args.no_cloudflare:
            cf_process = start_cloudflare()
            cf_pid = cf_process.pid
        
        console.print(get_server_table(servers, cf_pid))  
        console.print(Panel(
            "[yellow]Press Ctrl+C to stop servers. Use 'watch -n1 python main.py --status' to monitor.[/yellow]\n"
            "[cyan]API: https://bluppi-api.saikat.in/[/cyan]\n"
            "[cyan]WS1: wss://bluppi-ws1.saikat.in/[/cyan]\n"
            "[cyan]WS2: wss://bluppi-ws2.saikat.in/[/cyan]\n"
            "[cyan]gRPC: https://bluppi-grpc.saikat.in/[/cyan]"
        ))
        
        for server in servers:
            if server.process:
                server.process.wait()
        
    except KeyboardInterrupt:
        console.print("\n[bold red]Shutting down servers...[/bold red]")
        
        for server in servers:
            server.stop()
        
        if cf_pid:
            try:
                os.kill(cf_pid, signal.SIGINT)
            except Exception:
                pass 
        
        time.sleep(1)
        sys.exit(0)

if __name__ == "__main__":
    main()