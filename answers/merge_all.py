import json, glob, os

combined = {}
for f in sorted(glob.glob(os.path.join(os.path.dirname(__file__), 'COMBINED*.json'))):
    with open(f) as fh:
        data = json.load(fh)
    combined.update(data)
    print(f"{os.path.basename(f)}: {len(data)} entries")

out = os.path.join(os.path.dirname(__file__), 'MASTER.json')
with open(out, 'w') as fh:
    json.dump(combined, fh, indent=2)

print(f"\nTotal: {len(combined)} entries -> MASTER.json")