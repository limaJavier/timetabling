import json
import os

def toBool(matrix):
    for i in range(len(matrix)):
        for j in range(len(matrix[0])):
            matrix[i][j] = matrix[i][j] == 1
    return matrix

SOURCE = 'source'
OUT = 'out'

for satisfiability_status in ['satisfiable', 'unsatisfiable']:
    source = f'{SOURCE}/{satisfiability_status}'
    out = f'{OUT}/{satisfiability_status}'

    # Make sure that output directories exists
    if not os.path.exists(out):
        os.makedirs(out)

    for dir_name in os.listdir(source):
        metadata_professors_path = f'{source}/{dir_name}/metadata_professors.json'
        metadata_rooms_path = f'{source}/{dir_name}/metadata_rooms.json'
        metadata_classes_path = f'{source}/{dir_name}/metadata_classes.json'

        with open(metadata_professors_path, 'r') as file:
            professors = json.load(file)
        with open(metadata_rooms_path, 'r') as file:
            rooms = json.load(file)
        with open(metadata_classes_path, 'r') as file:
            classes = json.load(file)

        for i, professor in enumerate(professors):
            professor['id'] = i
        for i, room in enumerate(rooms):
            room['id'] = i
        for i, class_metadata in enumerate(classes):
            class_metadata['id'] = i

        entries = []
        for file_name in os.listdir(f"{source}/{dir_name}"):
            path = f"{source}/{dir_name}/{file_name}"
            if 'metadata' not in file_name:
                with open(path, 'r') as file:
                    config = json.load(file)

                major, year, type = config['major'], config['year'], config['type']

                for meta_entry in config['curriculum']:
                    # Adjust name to major, year and type
                    meta_entry['name'] += f"_{major}_{year}_{type}"
                        
                    # Adjust groups to account only for their index
                    groups = []
                    for raw_group in meta_entry['groups']:
                        group = []
                        for raw_class in raw_group:
                            class_name = raw_class
                            if all(map(lambda c: c.isnumeric(), raw_class)):
                                class_name = f'{major}{year}{raw_class}'
                            class_index = list(map(lambda class_metadata: class_metadata['name'], classes)).index(class_name)
                            group.append(class_index)
                        groups.append(group)
                    meta_entry['groups'] = groups

                    # Adjust professor to account only for its index 
                    meta_entry['professor'] = list(map(lambda professor_metadata: professor_metadata['name'], professors)).index(meta_entry['professor'])

                    # Adjust room to account only for its index 
                    indexed_rooms = []
                    for raw_room in meta_entry['rooms']:
                        room_index = list(map(lambda room_metadata: room_metadata['name'], rooms)).index(raw_room)
                        indexed_rooms.append(room_index)
                    meta_entry['rooms'] = indexed_rooms
                
                    # Adjust permissibility matrix to be a truly boolean one
                    meta_entry['permissibility'] = toBool(meta_entry['permissibility'])

                    for group in meta_entry['groups']:
                        entry = {}
                        entry['subject'] = meta_entry['name']
                        entry['professor'] = meta_entry['professor']
                        entry['classes'] = group
                        entry['lessons'] = meta_entry['lessons']
                        entry['permissibility'] = meta_entry['permissibility']
                        entry['rooms'] = meta_entry['rooms']
                        entries.append(entry)

        # Extract subjects
        subjects = []
        id = 0
        for entry in entries:
            subjectName = entry['subject']
            subject = list(filter(lambda subject: subject['name'] == subjectName, subjects))
            # Check if subject already exists
            if len(subject) == 0:
                subject = {'id': id, 'name': subjectName}
                subjects.append(subject)
                id += 1
            else:
                subject = subject[0]
            entry['subject'] = subject['id'] # Account only for subject's index

        periods = None
        days = None
        for i, entry in enumerate(entries):
            # Verify permissibility dimensions for all subject-professors
            permissibility = entry['permissibility']
            if periods != None and days != None:
                assert len(permissibility) == periods, True
                assert len(permissibility[0]) == days, True
            periods = len(permissibility)
            days = len(permissibility[0])

        instance = {
            'subjects': subjects,
            'professors': professors,
            'classes': classes,
            'rooms': rooms,
            'entries': entries,
        }

        with open(f'{out}/{dir_name}.json', '+w') as file:
            json.dump(instance, file) 
    
print('Done')
