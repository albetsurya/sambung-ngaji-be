function testArchiveFolder() {
  Logger.log("=== TEST ARCHIVE FOLDER ===");

  var mainFolder = getPhotoFolder_();
  var archiveFolder = getPhotoArchiveFolder_();

  Logger.log("Main folder:");
  Logger.log("  Name: " + mainFolder.getName());
  Logger.log("  ID: " + mainFolder.getId());
  Logger.log("  URL: " + mainFolder.getUrl());

  Logger.log("");
  Logger.log("Archive folder:");
  Logger.log("  Name: " + archiveFolder.getName());
  Logger.log("  ID: " + archiveFolder.getId());
  Logger.log("  URL: " + archiveFolder.getUrl());

  // Count files
  var mainCount = 0;
  var mainFiles = mainFolder.getFiles();
  while (mainFiles.hasNext()) { mainFiles.next(); mainCount++; }

  var archiveCount = 0;
  var archiveFiles = archiveFolder.getFiles();
  while (archiveFiles.hasNext()) { archiveFiles.next(); archiveCount++; }

  Logger.log("");
  Logger.log("Files:");
  Logger.log("  Main: " + mainCount);
  Logger.log("  Archive: " + archiveCount);
}