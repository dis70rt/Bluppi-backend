import time
from tqdm import tqdm

def demo_tracks():
    """Simulate tracks insertion with progress bar."""
    total = 150_000
    batch_size = 10_000
    
    with tqdm(total=total, desc="📀 Tracks", unit="row", colour="green") as pbar:
        processed = 0
        while processed < total:
            # Simulate batch processing delay
            time.sleep(0.3)
            batch = min(batch_size, total - processed)
            processed += batch
            pbar.update(batch)
    
    print(f"✅ Tracks complete: {total:,}")


def demo_artists():
    """Simulate artists insertion with progress bar."""
    total = 45_000
    batch_size = 10_000
    
    with tqdm(total=total, desc="🎤 Artists", unit="row", colour="cyan") as pbar:
        processed = 0
        while processed < total:
            time.sleep(0.2)
            batch = min(batch_size, total - processed)
            processed += batch
            pbar.update(batch)
    
    print(f"✅ Artists complete: {total:,}")


def demo_track_artists():
    """Simulate track_artists insertion with progress bar."""
    total = 200_000
    batch_size = 10_000
    
    with tqdm(total=total, desc="🔗 Relations", unit="row", colour="yellow") as pbar:
        processed = 0
        while processed < total:
            time.sleep(0.25)
            batch = min(batch_size, total - processed)
            processed += batch
            pbar.update(batch)
    
    print(f"✅ Track-Artists complete: {total:,}")


def main():
    print("🚀 Starting ETL Demo...\n")
    
    demo_tracks()
    print()
    
    demo_artists()
    print()
    
    demo_track_artists()
    
    print("\n🎉 ETL COMPLETE!")


if __name__ == "__main__":
    main()