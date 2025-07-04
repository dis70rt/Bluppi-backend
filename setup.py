import subprocess
import sys
import os
from rich.console import Console
from rich.panel import Panel
from rich.prompt import Confirm

console = Console()

def check_dependencies():
    required = ["uvicorn", "cloudflared"]
    missing = []
    
    for dep in required:
        try:
            subprocess.run(["which", dep], check=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        except subprocess.CalledProcessError:
            missing.append(dep)
    
    return missing

def setup():
    console.print(Panel.fit("[bold blue]Bluppi Setup Assistant[/bold blue]", subtitle="Setting up environment..."))
    
    missing_deps = check_dependencies()
    if missing_deps:
        console.print(f"[red]Missing dependencies: {', '.join(missing_deps)}[/red]")
        console.print("[yellow]Please install missing dependencies before continuing.[/yellow]")
        sys.exit(1)
    
    services = ["redis-server", "postgresql"]
    for service in services:
        try:
            subprocess.run(["systemctl", "is-active", "--quiet", service], check=True)
            console.print(f"[green]{service} is already running[/green]")
        except subprocess.CalledProcessError:
            console.print(f"[yellow]{service} is not running[/yellow]")
            if Confirm.ask(f"Do you want to start {service}?"):
                console.print(Panel.fit(
                    "[yellow]⚠️ Sudo privileges required to start system services[/yellow]\n\n"
                    "[white]Please enter your password when prompted.[/white]",
                    title="[bold blue]Sudo Required[/bold blue]", 
                    border_style="blue"
                ))
                try:
                    subprocess.run(["sudo", "systemctl", "start", service], check=True)
                    console.print(f"[green]{service} started successfully[/green]")
                except subprocess.CalledProcessError:
                    console.print(f"[red]Failed to start {service}[/red]")
    
    console.print("\n[green]Setup complete! You can now run 'python main.py'[/green]")

if __name__ == "__main__":
    setup()