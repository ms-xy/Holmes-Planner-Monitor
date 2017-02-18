package org.holmesprocessing.totem.monitoring.linux;

import java.io.FileReader;
import java.io.IOException;

/**
 * Created by nighty on 17.02.17.
 */
public class ProcFileReader {

    public static String read(String path) throws IOException {
        FileReader f = new FileReader(path);
        int l;
        char[] buf = new char[0x200];
        StringBuilder strbuilder = new StringBuilder();
        while ((l = f.read(buf, 0, buf.length)) > 0) {
            strbuilder.append(buf, 0, l);
        }
        f.close();
        return strbuilder.toString();
    }

}
