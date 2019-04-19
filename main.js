const fs = require('fs');
var path = require('path');

const filePatterns = ['.js', '.css', '.html', '.yml'];
const keywords = ['cache:'];
const matchedFilePaths = [];

Walk('artifact')
  .then(files => {
    files.forEach(file => {
      const matched = filePatterns.filter(e => e.indexOf('.js') > -1).length;

      if (matched) {
        const match = fs.readFileSync(file, 'utf-8').includes(keywords);

        if (match) {
          matchedFilePaths.push(file);
        }
      }
    });

    console.log({ matchedFilePaths });
  })
  .catch(err => {
    throw new Error(err);
  });

function Walk(dir) {
  return new Promise((resolve, reject) => {
    fs.readdir(dir, (error, files) => {
      if (error) {
        return reject(error);
      }

      Promise.all(
        files.map(file => {
          return new Promise((resolve, reject) => {
            const filepath = path.join(dir, file);
            fs.stat(filepath, (error, stats) => {
              if (error) {
                return reject(error);
              }

              if (stats.isDirectory()) {
                Walk(filepath).then(resolve);
              } else if (stats.isFile()) {
                resolve(filepath);
              }
            });
          });
        }),
      )
        .then(foldersContents => {
          resolve(
            foldersContents.reduce(
              (all, folderContents) => all.concat(folderContents),
              [],
            ),
          );
        })
        .catch(err => {
          throw new Error(err);
        });
    });
  });
}
